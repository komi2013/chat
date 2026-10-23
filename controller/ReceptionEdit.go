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

	"chat/common"
	"chat/table"
)

// ReceptionEdit saves reception/menu master data according to
// postgres_schema_optimized.sql. Master arrays are stored in reception JSONB;
// bookings, queues, facility codes, and facility occupancy remain operational rows.
// receptionEditPayload embeds the database model while accepting operational
// facility fields that are returned by the API but not stored in facilities JSONB.
type receptionEditPayload struct {
	table.ReceptionStruct
	Facilities []table.FacilityView `json:"facilities"`
}

type receptionMenuPayload struct {
	UpdatedAt   time.Time          `json:"updatedAt,omitempty"`
	Menus       []table.Menu       `json:"menus"`
	ItemDetails []table.ItemDetail `json:"itemDetails"`
}

func ReceptionEdit(w http.ResponseWriter, r *http.Request) {
	var reception receptionEditPayload
	if err := json.Unmarshal([]byte(r.FormValue("reception")), &reception); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON reception", http.StatusOK)
		return
	}

	var menu receptionMenuPayload
	if err := json.Unmarshal([]byte(r.FormValue("menu")), &menu); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON menu", http.StatusOK)
		return
	}

	aliasName := r.FormValue("aliasName")
	channelIDText := r.FormValue("channelID")
	channelID, err := strconv.ParseInt(channelIDText, 10, 64)
	if err != nil || channelID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid channelID", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var aliasExists bool
	err = common.DB.AliasDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM alias
			WHERE user_id = $1 AND alias_name = $2 AND channel_id = $3
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

	if reception.ChannelID == 0 {
		reception.ChannelID = channelID
	} else if reception.ChannelID != channelID {
		common.WriteResponseWithSession(w, session, "channelID does not match reception.channelID", http.StatusOK)
		return
	}

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin reception transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	isNewReception := reception.ReceptionID <= 0
	if isNewReception {
		if !containsString(reception.AdminNames, aliasName) {
			common.WriteResponseWithSession(w, session, "Only reception admins can create or edit this reception", http.StatusOK)
			return
		}
	} else {
		var storedChannelID int64
		var storedAdminNames []string
		err = tx.QueryRowContext(ctx, `
			SELECT channel_id, admin_names
			FROM reception
			WHERE reception_id = $1
			FOR UPDATE
		`, reception.ReceptionID).Scan(&storedChannelID, pq.Array(&storedAdminNames))
		if err != nil {
			if err == sql.ErrNoRows {
				common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
			} else {
				common.WriteResponseWithSession(w, session, "Failed to read reception permissions: "+err.Error(), http.StatusOK)
			}
			return
		}
		if storedChannelID != channelID {
			common.WriteResponseWithSession(w, session, "channelID does not match reception", http.StatusOK)
			return
		}
		if !containsString(storedAdminNames, aliasName) {
			logger := common.NewDailyLogger("reception_")
			logger.Printf("Unauthorized edit attempt: userID=%d aliasName=%s channelID=%d receptionID=%d", session.UserID, aliasName, channelID, reception.ReceptionID)
			common.WriteResponseWithSession(w, session, "You do not have permission to edit this reception", http.StatusOK)
			return
		}
	}

	// Save images only after authentication and authorization checks.
	for i := range menu.ItemDetails {
		if menu.ItemDetails[i].ImgPath == "" {
			continue
		}
		savedPath, saveErr := common.ImgSave(
			menu.ItemDetails[i].ImgPath,
			session.UserID,
			aliasName,
			channelIDText,
			3,
			6,
		)
		if saveErr != nil {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("ItemDetails[%d].ImgPath: %v", i, saveErr), http.StatusOK)
			return
		}
		menu.ItemDetails[i].ImgPath = savedPath
	}

	// Allocate stable IDs for new facilities and shifts before marshaling JSONB.
	for i := range reception.Facilities {
		if reception.Facilities[i].FacilityID <= 0 {
			if err := tx.QueryRowContext(ctx, `SELECT nextval('reception_facility_id_seq')`).Scan(&reception.Facilities[i].FacilityID); err != nil {
				common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to allocate facility ID at index %d: %v", i, err), http.StatusOK)
				return
			}
		}
	}
	reception.ReceptionStruct.Facilities = make([]table.FacilityMaster, len(reception.Facilities))
	for i, facility := range reception.Facilities {
		reception.ReceptionStruct.Facilities[i] = table.FacilityMaster{
			FacilityID: facility.FacilityID, FacilityName: facility.FacilityName,
			Capacity: facility.Capacity, Passcodes: facility.Passcodes, Bookable: facility.Bookable,
		}
	}
	shiftIDs := make(map[int64]struct{}, len(reception.Shifts))
	for i := range reception.Shifts {
		if reception.Shifts[i].ShiftID <= 0 {
			if err := tx.QueryRowContext(ctx, `SELECT nextval('reception_shift_id_seq')`).Scan(&reception.Shifts[i].ShiftID); err != nil {
				common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to allocate shift ID at index %d: %v", i, err), http.StatusOK)
				return
			}
		}
		if _, duplicate := shiftIDs[reception.Shifts[i].ShiftID]; duplicate {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("Duplicate shiftID %d", reception.Shifts[i].ShiftID), http.StatusOK)
			return
		}
		shiftIDs[reception.Shifts[i].ShiftID] = struct{}{}
	}
	for i, staff := range reception.WorkStaffs {
		if _, ok := shiftIDs[staff.ShiftID]; !ok {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("workStaff[%d] references unknown shiftID %d", i, staff.ShiftID), http.StatusOK)
			return
		}
	}

	// Facility currentCode and BookTimes are deliberately excluded from the
	// facilities JSONB column. Current codes are stored in the registry table;
	// booking occupancy remains in reception_facility_book_time.
	facilitiesJSON, err := marshalFacilitiesMaster(reception.Facilities)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode facilities: "+err.Error(), http.StatusOK)
		return
	}
	passcodesJSON, err := json.Marshal(nonNilValue(reception.Passcodes))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode passcodes: "+err.Error(), http.StatusOK)
		return
	}
	asksJSON, err := json.Marshal(nonNilValue(reception.Asks))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode asks: "+err.Error(), http.StatusOK)
		return
	}
	askChoicesJSON, err := json.Marshal(nonNilValue(reception.AskChoices))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode ask choices: "+err.Error(), http.StatusOK)
		return
	}
	askMultiChoicesJSON, err := json.Marshal(nonNilValue(reception.AskMultiChoices))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode ask multi-choices: "+err.Error(), http.StatusOK)
		return
	}
	openTimesJSON, err := json.Marshal(nonNilValue(reception.OpenTimes))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode open times: "+err.Error(), http.StatusOK)
		return
	}
	shiftsJSON, err := json.Marshal(nonNilValue(reception.Shifts))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode shifts: "+err.Error(), http.StatusOK)
		return
	}
	staffSkillsJSON, err := json.Marshal(nonNilValue(reception.StaffSkills))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode staff skills: "+err.Error(), http.StatusOK)
		return
	}
	workStaffsJSON, err := json.Marshal(nonNilValue(reception.WorkStaffs))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode work staffs: "+err.Error(), http.StatusOK)
		return
	}
	waitConfigsJSON, err := json.Marshal(nonNilValue(reception.WaitConfigs))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode wait configs: "+err.Error(), http.StatusOK)
		return
	}
	menusJSON, err := json.Marshal(nonNilValue(menu.Menus))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode menus: "+err.Error(), http.StatusOK)
		return
	}
	itemDetailsJSON, err := json.Marshal(nonNilValue(menu.ItemDetails))
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to encode item details: "+err.Error(), http.StatusOK)
		return
	}

	now := time.Now()
	reception.UpdatedAt = now
	menu.UpdatedAt = now
	reception.ChannelID = channelID

	if isNewReception {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO reception (
				channel_id, updated_at, reception_title, work_staff_need,
				admin_names, join_names, inquiry_names, subscriptions,
				order_user_ids, skills,
				passcodes, asks, ask_choices, ask_multi_choices, facilities,
				open_times, shifts, staff_skills, work_staffs, wait_configs,
				menus, item_details
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11::jsonb, $12::jsonb, $13::jsonb, $14::jsonb, $15::jsonb,
				$16::jsonb, $17::jsonb, $18::jsonb, $19::jsonb, $20::jsonb,
				$21::jsonb, $22::jsonb
			)
			RETURNING reception_id
		`,
			reception.ChannelID, reception.UpdatedAt, reception.ReceptionTitle, reception.WorkStaffNeed,
			pq.Array(nonNilStrings(reception.AdminNames)), pq.Array(nonNilStrings(reception.JoinNames)),
			pq.Array(nonNilStrings(reception.InquiryNames)), pq.Array(nonNilStrings(reception.Subscriptions)),
			pq.Array(nonNilInt64s(reception.OrderUserIDs)), pq.Array(nonNilStrings(reception.Skills)),
			string(passcodesJSON), string(asksJSON), string(askChoicesJSON), string(askMultiChoicesJSON), string(facilitiesJSON),
			string(openTimesJSON), string(shiftsJSON), string(staffSkillsJSON), string(workStaffsJSON), string(waitConfigsJSON),
			string(menusJSON), string(itemDetailsJSON),
		).Scan(&reception.ReceptionID)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to insert reception: "+err.Error(), http.StatusOK)
			return
		}
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE reception SET
				channel_id = $1, updated_at = $2, reception_title = $3, work_staff_need = $4,
				admin_names = $5, join_names = $6, inquiry_names = $7, subscriptions = $8,
				order_user_ids = $9, skills = $10,
				passcodes = $11::jsonb, asks = $12::jsonb, ask_choices = $13::jsonb,
				ask_multi_choices = $14::jsonb, facilities = $15::jsonb, open_times = $16::jsonb,
				shifts = $17::jsonb, staff_skills = $18::jsonb, work_staffs = $19::jsonb,
				wait_configs = $20::jsonb, menus = $21::jsonb, item_details = $22::jsonb
			WHERE reception_id = $23
		`,
			reception.ChannelID, reception.UpdatedAt, reception.ReceptionTitle, reception.WorkStaffNeed,
			pq.Array(nonNilStrings(reception.AdminNames)), pq.Array(nonNilStrings(reception.JoinNames)),
			pq.Array(nonNilStrings(reception.InquiryNames)), pq.Array(nonNilStrings(reception.Subscriptions)),
			pq.Array(nonNilInt64s(reception.OrderUserIDs)), pq.Array(nonNilStrings(reception.Skills)),
			string(passcodesJSON), string(asksJSON), string(askChoicesJSON), string(askMultiChoicesJSON),
			string(facilitiesJSON), string(openTimesJSON), string(shiftsJSON), string(staffSkillsJSON),
			string(workStaffsJSON), string(waitConfigsJSON), string(menusJSON), string(itemDetailsJSON), reception.ReceptionID,
		)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to update reception: "+err.Error(), http.StatusOK)
			return
		}
	}

	// Sync the active facility-code registry only. Do not delete booking,
	// queue, or facility-occupancy rows as a side effect of editing master data.
	if _, err = tx.ExecContext(ctx, `DELETE FROM reception_facility_code WHERE reception_id = $1`, reception.ReceptionID); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to replace facility codes: "+err.Error(), http.StatusOK)
		return
	}
	for i, facility := range reception.Facilities {
		if facility.CurrentCode == "" {
			continue
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO reception_facility_code (current_code, reception_id, facility_id, updated_at)
			VALUES ($1, $2, $3, $4)
		`, facility.CurrentCode, reception.ReceptionID, facility.FacilityID, now)
		if err != nil {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to save facilities[%d].currentCode: %v", i, err), http.StatusOK)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit reception transaction: "+err.Error(), http.StatusOK)
		return
	}

	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception.ReceptionStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		return
	}
}

func marshalFacilitiesMaster[T any](facilities []T) (json.RawMessage, error) {
	// Round-trip through generic JSON maps so API-only/operational fields do not
	// accidentally get persisted in reception.facilities JSONB.
	if facilities == nil {
		facilities = []T{}
	}
	raw, err := json.Marshal(facilities)
	if err != nil {
		return nil, err
	}
	var objects []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &objects); err != nil {
		return nil, err
	}
	for _, object := range objects {
		delete(object, "currentCode")
		delete(object, "bookTimes")
	}
	return json.Marshal(objects)
}

func nonNilValue[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nonNilInt64s(values []int64) []int64 {
	if values == nil {
		return []int64{}
	}
	return values
}
