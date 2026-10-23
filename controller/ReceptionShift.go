package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/collection"
	"chat/common"
)

// workStaffShiftRow carries the database relationship to shift_id internally.
// ShiftID is not required in the public WorkStaff JSON payload.
type workStaffShiftRow struct {
	ID        int64
	ShiftID   int64
	SortOrder int
	Staff     collection.WorkStaff
}

func ReceptionShift(w http.ResponseWriter, r *http.Request) {
	aliasName := r.FormValue("aliasName")
	channelID, err := strconv.ParseInt(r.FormValue("channelID"), 10, 64)
	if err != nil || channelID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid channelID", http.StatusOK)
		return
	}
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID", http.StatusOK)
		return
	}

	var updatedShifts []collection.Shift
	if err := json.Unmarshal([]byte(r.FormValue("updatedShifts")), &updatedShifts); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "updatedShifts JSON Unmarshal Error", http.StatusOK)
		return
	}

	// Every update must target an existing shift. Open must never be negative.
	seenShiftIDs := make(map[int64]bool, len(updatedShifts))
	for i, shift := range updatedShifts {
		if shift.ShiftID <= 0 {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), fmt.Sprintf("updatedShifts[%d]: invalid shiftID", i), http.StatusOK)
			return
		}
		if seenShiftIDs[shift.ShiftID] {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), fmt.Sprintf("updatedShifts[%d]: duplicate shiftID", i), http.StatusOK)
			return
		}
		seenShiftIDs[shift.ShiftID] = true
		if shift.Open < 0 {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), fmt.Sprintf("updatedShifts[%d]: open must be >= 0", i), http.StatusOK)
			return
		}
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validate the posted user/alias/channel tuple against the alias table.
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

	// loadReception reads the normalized child tables. Its shift query must select
	// shift_id (not the old generic id column) and assign it to collection.Shift.ShiftID.
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
			common.WriteResponseWithSession(w, session, "A shift has an invalid shiftID; check the reception_get.go loader and schema", http.StatusOK)
			return
		}
		shiftPositions[shift.ShiftID] = i
	}

	workRows, err := loadReceptionWorkStaffRows(ctx, tx, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load work staff: "+err.Error(), http.StatusOK)
		return
	}

	// Apply only the shifts supplied by the client. ShiftStart and Skill may change
	// because shiftID, rather than those mutable values, identifies the row.
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
		for _, row := range workRows {
			if row.ShiftID == updated.ShiftID {
				existingAliasSet[row.Staff.AliasName] = true
			}
		}

		// Remove staff no longer within the Open portion of AliasNames;
		// update retained staff's working times when ShiftStart/ShiftEnd changes.
		filteredRows := workRows[:0]
		for _, row := range workRows {
			if row.ShiftID != updated.ShiftID {
				filteredRows = append(filteredRows, row)
				continue
			}

			if !openAliasSet[row.Staff.AliasName] {
				if _, err := tx.ExecContext(ctx, `
					DELETE FROM reception_work_staff
					WHERE id = $1 AND reception_id = $2 AND shift_id = $3
				`, row.ID, receptionID, updated.ShiftID); err != nil {
					common.WriteResponseWithSession(w, session, "Failed to remove work staff: "+err.Error(), http.StatusOK)
					return
				}
				continue
			}

			if _, err := tx.ExecContext(ctx, `
				UPDATE reception_work_staff
				SET work_start = $1, work_end = $2
				WHERE id = $3 AND reception_id = $4 AND shift_id = $5
			`, updated.ShiftStart, updated.ShiftEnd, row.ID, receptionID, updated.ShiftID); err != nil {
				common.WriteResponseWithSession(w, session, "Failed to update work staff: "+err.Error(), http.StatusOK)
				return
			}
			row.Staff.WorkStart = updated.ShiftStart
			row.Staff.WorkEnd = updated.ShiftEnd
			filteredRows = append(filteredRows, row)
		}
		workRows = filteredRows

		// Add newly opened aliases. Open greater than len(AliasNames) means all
		// available aliases, matching the existing helper's behavior.
		nextSortOrder := 0
		for _, row := range workRows {
			if row.SortOrder >= nextSortOrder {
				nextSortOrder = row.SortOrder + 1
			}
		}
		for i, name := range updated.AliasNames {
			if i >= updated.Open {
				break
			}
			if existingAliasSet[name] {
				continue
			}

			seq := len(workRows) + 1
			var id int64
			err := tx.QueryRowContext(ctx, `
				INSERT INTO reception_work_staff (
					reception_id, shift_id, sort_order, alias_name, work_start, work_end, seq
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, receptionID, updated.ShiftID, nextSortOrder, name, updated.ShiftStart, updated.ShiftEnd, seq).Scan(&id)
			if err != nil {
				common.WriteResponseWithSession(w, session, "Failed to add work staff: "+err.Error(), http.StatusOK)
				return
			}
			workRows = append(workRows, workStaffShiftRow{
				ID:        id,
				ShiftID:   updated.ShiftID,
				SortOrder: nextSortOrder,
				Staff: collection.WorkStaff{
					ShiftID:   updated.ShiftID,
					AliasName: name,
					WorkStart: updated.ShiftStart,
					WorkEnd:   updated.ShiftEnd,
					Seq:       seq,
				},
			})
			nextSortOrder++
			existingAliasSet[name] = true
		}

		updated.AliasNames = nonNilShiftAliases(updated.AliasNames)
		result, err := tx.ExecContext(ctx, `
			UPDATE reception_shift
			SET alias_names = $1,
			    shift_start = $2,
			    shift_end = $3,
			    open = $4,
			    skill = $5,
			    fix = $6
			WHERE reception_id = $7 AND shift_id = $8
		`, pq.Array(updated.AliasNames), updated.ShiftStart, updated.ShiftEnd, updated.Open, updated.Skill, updated.Fix, receptionID, updated.ShiftID)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to update shift: "+err.Error(), http.StatusOK)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected != 1 {
			if err != nil {
				common.WriteResponseWithSession(w, session, "Failed to check updated shift: "+err.Error(), http.StatusOK)
			} else {
				common.WriteResponseWithSession(w, session, "Shift not found", http.StatusOK)
			}
			return
		}
		reception.Shifts[shiftIndex] = updated
	}

	// Remove expired shifts and their WorkStaffs using Japan local time.
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("JST", 9*60*60)
	}
	now := time.Now().In(jst)
	validShifts := make([]collection.Shift, 0, len(reception.Shifts))
	expiredShiftIDs := make(map[int64]bool)
	for _, shift := range reception.Shifts {
		shiftEndTime, parseErr := time.ParseInLocation("2006-01-02T15:04", shift.ShiftEnd, jst)
		if parseErr != nil || shiftEndTime.After(now) {
			validShifts = append(validShifts, shift)
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			DELETE FROM reception_work_staff
			WHERE reception_id = $1 AND shift_id = $2
		`, receptionID, shift.ShiftID); err != nil {
			common.WriteResponseWithSession(w, session, "Failed to delete expired work staff: "+err.Error(), http.StatusOK)
			return
		}
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM reception_shift
			WHERE reception_id = $1 AND shift_id = $2
		`, receptionID, shift.ShiftID); err != nil {
			common.WriteResponseWithSession(w, session, "Failed to delete expired shift: "+err.Error(), http.StatusOK)
			return
		}
		expiredShiftIDs[shift.ShiftID] = true
	}
	reception.Shifts = validShifts

	// Rebuild the response's WorkStaffs slice from the rows after all changes.
	reception.WorkStaffs = make([]collection.WorkStaff, 0, len(workRows))
	for _, row := range workRows {
		if expiredShiftIDs[row.ShiftID] {
			continue
		}
		reception.WorkStaffs = append(reception.WorkStaffs, row.Staff)
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit shift update: "+err.Error(), http.StatusOK)
		return
	}

	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		return
	}
}

func loadReceptionWorkStaffRows(ctx context.Context, q receptionQueryer, receptionID int64) ([]workStaffShiftRow, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, shift_id, sort_order, COALESCE(alias_name, ''),
		       COALESCE(work_start, ''), COALESCE(work_end, ''), seq
		FROM reception_work_staff
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []workStaffShiftRow
	for rows.Next() {
		var row workStaffShiftRow
		if err := rows.Scan(
			&row.ID,
			&row.ShiftID,
			&row.SortOrder,
			&row.Staff.AliasName,
			&row.Staff.WorkStart,
			&row.Staff.WorkEnd,
			&row.Staff.Seq,
		); err != nil {
			return nil, err
		}
		row.Staff.ShiftID = row.ShiftID
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func nonNilShiftAliases(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// getOpenAliasesFromShift returns the first Open aliases; a negative Open is
// rejected before this helper is called.
func getOpenAliasesFromShift(aliasNames []string, open int) []string {
	if open <= 0 {
		return []string{}
	}
	if len(aliasNames) > open {
		return aliasNames[:open]
	}
	return aliasNames
}
