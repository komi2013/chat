package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chat/common"
	"chat/table"
)

func ReceptionShift(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")
	aliasName := r.FormValue("aliasName")

	channelID, err := strconv.ParseInt(r.FormValue("channelID"), 10, 64)
	if err != nil || channelID <= 0 {
		common.WriteResponseWithoutSession(w, csrf, "Invalid channelID", http.StatusOK)
		return
	}
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, csrf, "Invalid receptionID", http.StatusOK)
		return
	}

	var updatedShifts []table.Shift
	if err := json.Unmarshal([]byte(r.FormValue("updatedShifts")), &updatedShifts); err != nil {
		common.WriteResponseWithoutSession(w, csrf, "updatedShifts JSON Unmarshal Error", http.StatusOK)
		return
	}

	seenShiftIDs := make(map[int64]bool, len(updatedShifts))
	for i, shift := range updatedShifts {
		if shift.ShiftID <= 0 {
			common.WriteResponseWithoutSession(w, csrf, fmt.Sprintf("updatedShifts[%d]: invalid shiftID", i), http.StatusOK)
			return
		}
		if seenShiftIDs[shift.ShiftID] {
			common.WriteResponseWithoutSession(w, csrf, fmt.Sprintf("updatedShifts[%d]: duplicate shiftID", i), http.StatusOK)
			return
		}
		seenShiftIDs[shift.ShiftID] = true
		if shift.Open < 0 {
			common.WriteResponseWithoutSession(w, csrf, fmt.Sprintf("updatedShifts[%d]: open must be >= 0", i), http.StatusOK)
			return
		}
	}

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var aliasExists bool
	err = common.DB.AliasDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE user_id = $1
			  AND alias_name = $2
			  AND channel_id = $3
		)
	`, session.UserID, aliasName, channelID).Scan(&aliasExists)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to validate alias: "+err.Error(), http.StatusOK)
		return
	}
	if !aliasExists {
		common.WriteResponseWithSession(w, session, "aliasName or channelID does not match this user", http.StatusOK)
		return
	}

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	// Lock the single reception master row while editing its JSONB arrays.
	var lockedReceptionID int64
	err = tx.QueryRowContext(ctx, `
		SELECT reception_id
		FROM reception
		WHERE reception_id = $1 AND channel_id = $2
		FOR UPDATE
	`, receptionID, channelID).Scan(&lockedReceptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "reception not found or channelID does not match reception", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to lock reception: "+err.Error(), http.StatusOK)
		}
		return
	}

	reception, err := loadReception(ctx, tx, receptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "reception not found", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		}
		return
	}
	if reception.ChannelID != channelID {
		common.WriteResponseWithSession(w, session, "channelID does not match reception", http.StatusOK)
		return
	}

	shiftPositions := make(map[int64]int, len(reception.Shifts))
	for i, shift := range reception.Shifts {
		if shift.ShiftID <= 0 {
			common.WriteResponseWithSession(w, session, "A shift has an invalid shiftID", http.StatusOK)
			return
		}
		if _, duplicate := shiftPositions[shift.ShiftID]; duplicate {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("Duplicate stored shiftID %d", shift.ShiftID), http.StatusOK)
			return
		}
		shiftPositions[shift.ShiftID] = i
	}

	for _, updated := range updatedShifts {
		shiftIndex, exists := shiftPositions[updated.ShiftID]
		if !exists {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("shiftID %d does not belong to this reception", updated.ShiftID), http.StatusOK)
			return
		}

		openAliases := getOpenAliasesFromShift(updated.AliasNames, updated.Open)
		openAliasSet := make(map[string]bool, len(openAliases))
		for _, name := range openAliases {
			openAliasSet[name] = true
		}

		existingAliasSet := make(map[string]bool)
		for _, staff := range reception.WorkStaffs {
			if staff.ShiftID == updated.ShiftID {
				existingAliasSet[staff.AliasName] = true
			}
		}

		// Keep only staff within the open portion of AliasNames and update their times.
		filteredStaff := make([]table.WorkStaff, 0, len(reception.WorkStaffs))
		for _, staff := range reception.WorkStaffs {
			if staff.ShiftID != updated.ShiftID {
				filteredStaff = append(filteredStaff, staff)
				continue
			}
			if !openAliasSet[staff.AliasName] {
				continue
			}
			staff.WorkStart = updated.ShiftStart
			staff.WorkEnd = updated.ShiftEnd
			filteredStaff = append(filteredStaff, staff)
		}
		reception.WorkStaffs = filteredStaff

		// Add aliases newly included in the open portion.
		for _, name := range openAliases {
			if existingAliasSet[name] {
				continue
			}
			reception.WorkStaffs = append(reception.WorkStaffs, table.WorkStaff{
				ShiftID:   updated.ShiftID,
				AliasName: name,
				WorkStart: updated.ShiftStart,
				WorkEnd:   updated.ShiftEnd,
				Seq:       len(reception.WorkStaffs) + 1,
			})
			existingAliasSet[name] = true
		}

		updated.AliasNames = nonNilShiftAliases(updated.AliasNames)
		reception.Shifts[shiftIndex] = updated
	}

	// Remove expired shifts and their associated work staff using Japan local time.
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("JST", 9*60*60)
	}
	now := time.Now().In(jst)
	validShifts := make([]table.Shift, 0, len(reception.Shifts))
	expiredShiftIDs := make(map[int64]bool)
	for _, shift := range reception.Shifts {
		shiftEndTime, parseErr := time.ParseInLocation("2006-01-02T15:04", shift.ShiftEnd, jst)
		if parseErr != nil || shiftEndTime.After(now) {
			validShifts = append(validShifts, shift)
			continue
		}
		expiredShiftIDs[shift.ShiftID] = true
	}
	reception.Shifts = validShifts

	validWorkStaffs := make([]table.WorkStaff, 0, len(reception.WorkStaffs))
	for _, staff := range reception.WorkStaffs {
		if !expiredShiftIDs[staff.ShiftID] {
			validWorkStaffs = append(validWorkStaffs, staff)
		}
	}
	reception.WorkStaffs = validWorkStaffs

	shiftsJSON, err := json.Marshal(nonNilShifts(reception.Shifts))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode shifts: "+err.Error(), http.StatusOK)
		return
	}
	workStaffsJSON, err := json.Marshal(nonNilWorkStaffs(reception.WorkStaffs))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode work staff: "+err.Error(), http.StatusOK)
		return
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE reception
		SET shifts = $1::jsonb,
		    work_staffs = $2::jsonb,
		    updated_at = CURRENT_TIMESTAMP
		WHERE reception_id = $3 AND channel_id = $4
	`, string(shiftsJSON), string(workStaffsJSON), receptionID, channelID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to save shifts: "+err.Error(), http.StatusOK)
		return
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit shift update: "+err.Error(), http.StatusOK)
		return
	}

	reception.UpdatedAt = time.Now()
	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func nonNilShiftAliases(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func getOpenAliasesFromShift(aliasNames []string, open int) []string {
	if open <= 0 {
		return []string{}
	}
	if len(aliasNames) > open {
		return aliasNames[:open]
	}
	return aliasNames
}

func nonNilShifts(values []table.Shift) []table.Shift {
	if values == nil {
		return []table.Shift{}
	}
	return values
}

func nonNilWorkStaffs(values []table.WorkStaff) []table.WorkStaff {
	if values == nil {
		return []table.WorkStaff{}
	}
	return values
}
