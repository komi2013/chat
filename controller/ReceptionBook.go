package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/common"
	"chat/table"
)

const receptionBookDateTimeLayout = "2006-01-02T15:04"

type receptionBookAppointment struct {
	BookID     int64
	Start      time.Time
	End        time.Time
	NeedSkill  string
	Candidates []string
}

func ReceptionBook(w http.ResponseWriter, r *http.Request) {
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID format", http.StatusOK)
		return
	}

	menuID, err := strconv.ParseInt(r.FormValue("menuID"), 10, 64)
	if err != nil || menuID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid menuID format", http.StatusOK)
		return
	}

	var answers []string
	if err := json.Unmarshal([]byte(r.FormValue("answers")), &answers); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON answers", http.StatusOK)
		return
	}
	if answers == nil {
		answers = []string{}
	}

	bookStartText := r.FormValue("bookStart")
	bookEndText := r.FormValue("bookEnd")
	if bookStartText == "" || bookEndText == "" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Missing bookStart or bookEnd", http.StatusOK)
		return
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// Fallback for containers without the timezone database.
		jst = time.FixedZone("JST", 9*60*60)
	}
	bookStartTime, err := time.ParseInLocation(receptionBookDateTimeLayout, bookStartText, jst)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid bookStart format", http.StatusOK)
		return
	}
	bookEndTime, err := time.ParseInLocation(receptionBookDateTimeLayout, bookEndText, jst)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid bookEnd format", http.StatusOK)
		return
	}
	if !bookEndTime.After(bookStartTime) {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "bookEnd must be later than bookStart", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}
	if session.Mail == "" || session.Telephone == "" {
		common.WriteResponseWithSession(w, session, "電話とメールを登録してください", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin booking transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	// Load the reception row and decode its JSONB master fields.
	var reception table.ReceptionStruct
	var passcodesJSON, asksJSON, askChoicesJSON, askMultiChoicesJSON []byte
	var facilitiesJSON, openTimesJSON, shiftsJSON, staffSkillsJSON []byte
	var workStaffsJSON, waitConfigsJSON, menusJSON, itemDetailsJSON []byte

	err = tx.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, admin_names, join_names, inquiry_names,
		       order_user_ids, updated_at, COALESCE(reception_title, ''),
		       skills, work_staff_need, passcodes, asks, ask_choices, ask_multi_choices,
		       facilities, open_times, shifts, staff_skills, work_staffs, wait_configs,
		       menus, item_details
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID, &reception.ChannelID,
		pq.Array(&reception.AdminNames), pq.Array(&reception.JoinNames),
		pq.Array(&reception.InquiryNames),
		pq.Array(&reception.OrderUserIDs), &reception.UpdatedAt,
		&reception.ReceptionTitle, pq.Array(&reception.Skills), &reception.WorkStaffNeed,
		&passcodesJSON, &asksJSON, &askChoicesJSON, &askMultiChoicesJSON,
		&facilitiesJSON, &openTimesJSON, &shiftsJSON, &staffSkillsJSON,
		&workStaffsJSON, &waitConfigsJSON, &menusJSON, &itemDetailsJSON,
	)
	if err == nil {
		for _, item := range []struct {
			name string
			raw  []byte
			dst  any
		}{
			{"passcodes", passcodesJSON, &reception.Passcodes},
			{"asks", asksJSON, &reception.Asks},
			{"ask_choices", askChoicesJSON, &reception.AskChoices},
			{"ask_multi_choices", askMultiChoicesJSON, &reception.AskMultiChoices},
			{"facilities", facilitiesJSON, &reception.Facilities},
			{"open_times", openTimesJSON, &reception.OpenTimes},
			{"shifts", shiftsJSON, &reception.Shifts},
			{"staff_skills", staffSkillsJSON, &reception.StaffSkills},
			{"work_staffs", workStaffsJSON, &reception.WorkStaffs},
			{"wait_configs", waitConfigsJSON, &reception.WaitConfigs},
			{"menus", menusJSON, &reception.Menus},
			{"item_details", itemDetailsJSON, &reception.ItemDetails},
		} {
			if len(item.raw) > 0 {
				if unmarshalErr := json.Unmarshal(item.raw, item.dst); unmarshalErr != nil {
					err = fmt.Errorf("decode reception.%s: %w", item.name, unmarshalErr)
					break
				}
			}
		}
	}
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "Failed to find reception", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to find reception: "+err.Error(), http.StatusOK)
		}
		return
	}

	// Menus are JSONB master data in reception.menus, not a separate table.
	var selectedMenu table.Menu
	menuFound := false
	for _, item := range reception.Menus {
		if item.MenuID == menuID {
			selectedMenu = item
			menuFound = true
			break
		}
	}
	if !menuFound {
		common.WriteResponseWithSession(w, session, "Service not found", http.StatusOK)
		return
	}

	// Facilities are JSONB master data in reception.facilities. Active codes and
	// occupied times remain in their dedicated operational tables.
	var facility table.FacilityMaster
	hasFacility := selectedMenu.NeedFacilityID > 0
	if hasFacility {
		facilityFound := false
		for _, item := range reception.Facilities {
			if item.FacilityID == selectedMenu.NeedFacilityID && item.Bookable {
				facility = item
				facilityFound = true
				break
			}
		}
		if !facilityFound {
			common.WriteResponseWithSession(w, session, "Facility not available", http.StatusOK)
			return
		}
		if facility.Capacity <= 0 {
			common.WriteResponseWithSession(w, session, "Facility not available", http.StatusOK)
			return
		}

		facilityTimes := make([]table.ReceptionFacilityBookTimeStruct, 0)
		rows, queryErr := tx.QueryContext(ctx, `
			SELECT facility_book_time_id, reception_id, facility_id, book_id,
			       sort_order, book_start, book_end
			FROM reception_facility_book_time
			WHERE reception_id = $1 AND facility_id = $2
			ORDER BY sort_order, facility_book_time_id
		`, receptionID, facility.FacilityID)
		if queryErr != nil {
			common.WriteResponseWithSession(w, session, "Failed to load facility booking times: "+queryErr.Error(), http.StatusOK)
			return
		}
		for rows.Next() {
			var item table.ReceptionFacilityBookTimeStruct
			if err := rows.Scan(&item.ID, &item.ReceptionID, &item.FacilityID, &item.BookID, &item.SortOrder, &item.BookStart, &item.BookEnd); err != nil {
				rows.Close()
				common.WriteResponseWithSession(w, session, "Failed to read facility booking times: "+err.Error(), http.StatusOK)
				return
			}
			facilityTimes = append(facilityTimes, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			common.WriteResponseWithSession(w, session, "Failed to read facility booking times: "+err.Error(), http.StatusOK)
			return
		}
		rows.Close()

		overlappingFacilityBookings := 0
		for _, item := range facilityTimes {
			start, startErr := receptionBookParseJST(item.BookStart, jst)
			end, endErr := receptionBookParseJST(item.BookEnd, jst)
			if startErr != nil || endErr != nil {
				// Malformed stored data should not be counted as valid capacity usage.
				continue
			}
			if intervalsOverlap(bookStartTime, bookEndTime, start, end) {
				overlappingFacilityBookings++
			}
		}
		if overlappingFacilityBookings >= facility.Capacity {
			common.WriteResponseWithSession(w, session, "Facility not available", http.StatusOK)
			return
		}
	}

	// Staff checks are only required when WorkStaffNeed is enabled.
	// NeedSkill selects which staff are eligible; an empty NeedSkill permits any
	// staff member whose work interval covers the booking.
	if reception.WorkStaffNeed {
		workStaffs := reception.WorkStaffs
		staffSkills := reception.StaffSkills
		menuSkills := make(map[int64]string, len(reception.Menus))
		for _, item := range reception.Menus {
			menuSkills[item.MenuID] = item.NeedSkill
		}
		bookings, err := receptionBookLoadBookings(ctx, tx, receptionID)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to load existing bookings: "+err.Error(), http.StatusOK)
			return
		}

		newAppointment := receptionBookAppointment{
			BookID:     0, // New booking; generated after all validations.
			Start:      bookStartTime,
			End:        bookEndTime,
			NeedSkill:  selectedMenu.NeedSkill,
			Candidates: receptionBookEligibleStaff(bookStartTime, bookEndTime, selectedMenu.NeedSkill, workStaffs, staffSkills, jst),
		}
		if len(newAppointment.Candidates) == 0 {
			common.WriteResponseWithSession(w, session, "No available staff", http.StatusOK)
			return
		}

		appointments := make([]receptionBookAppointment, 0, len(bookings)+1)
		for _, booking := range bookings {
			start, startErr := receptionBookParseJST(booking.BookStart, jst)
			end, endErr := receptionBookParseJST(booking.BookEnd, jst)
			if startErr != nil || endErr != nil || !end.After(start) {
				// Invalid legacy appointments cannot safely be used for overlap matching.
				continue
			}
			needSkill := menuSkills[booking.MenuID]
			appointments = append(appointments, receptionBookAppointment{
				BookID:     booking.BookID,
				Start:      start,
				End:        end,
				NeedSkill:  needSkill,
				Candidates: receptionBookEligibleStaff(start, end, needSkill, workStaffs, staffSkills, jst),
			})
		}

		// Include the new appointment and find all existing appointments connected
		// to it by overlapping intervals. A staff alias cannot be assigned to two
		// overlapping appointments in this connected set.
		newAppointmentIndex := len(appointments)
		appointments = append(appointments, newAppointment)
		component := receptionBookOverlapComponent(appointments, newAppointmentIndex)
		if !receptionBookCanAssignStaff(component) {
			common.WriteResponseWithSession(w, session, "this time already booked", http.StatusOK)
			return
		}
	}

	bookStart := bookStartTime.Format(receptionBookDateTimeLayout)
	bookEnd := bookEndTime.Format(receptionBookDateTimeLayout)
	createdAt := time.Now().In(jst)

	var bookID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO reception_book (
			reception_id, sort_order, book_start, book_end, answers,
			menu_id, people, created_at
		)
		VALUES (
			$1,
			(SELECT COALESCE(MAX(sort_order) + 1, 0) FROM reception_book WHERE reception_id = $1),
			$2, $3, $4, $5, 0, $6
		)
		RETURNING book_id
	`, receptionID, bookStart, bookEnd, pq.Array(answers), menuID, createdAt).Scan(&bookID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to save booking: "+err.Error(), http.StatusOK)
		return
	}

	// Keep the facility schedule table so ReceptionGet can load facility times
	// directly. It is updated atomically with reception_book.
	if hasFacility {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO reception_facility_book_time (
				reception_id, facility_id, book_id, sort_order, book_start, book_end
			)
			VALUES (
				$1, $2, $3,
				(SELECT COALESCE(MAX(sort_order) + 1, 0)
				 FROM reception_facility_book_time WHERE reception_id = $1 AND facility_id = $2),
				$4, $5
			)
		`, receptionID, facility.FacilityID, bookID, bookStart, bookEnd)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to update facility booking time: "+err.Error(), http.StatusOK)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit booking: "+err.Error(), http.StatusOK)
		return
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		BookID       int64    `json:"bookID"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		BookID:       bookID,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func receptionBookLoadBookings(
	ctx context.Context,
	tx *sql.Tx,
	receptionID int64,
) ([]table.ReceptionBookStruct, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT book_id, reception_id, sort_order, book_start, book_end, answers,
		       COALESCE(menu_id, 0), COALESCE(people, 0), COALESCE(created_at, CURRENT_TIMESTAMP)
		FROM reception_book
		WHERE reception_id = $1
		ORDER BY sort_order, book_id
	`, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]table.ReceptionBookStruct, 0)
	for rows.Next() {
		var item table.ReceptionBookStruct
		if err := rows.Scan(&item.BookID, &item.ReceptionID, &item.SortOrder,
			&item.BookStart, &item.BookEnd, pq.Array(&item.Answers),
			&item.MenuID, &item.People, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func receptionBookParseJST(value string, jst *time.Location) (time.Time, error) {
	return time.ParseInLocation(receptionBookDateTimeLayout, value, jst)
}

func receptionBookStaffHasSkill(
	aliasName string,
	needSkill string,
	staffSkills []table.StaffSkill,
) bool {
	for _, staff := range staffSkills {
		if staff.AliasName != aliasName {
			continue
		}
		for _, skill := range staff.Skills {
			if skill == needSkill {
				return true
			}
		}
	}
	return false
}

func receptionBookEligibleStaff(
	start time.Time,
	end time.Time,
	needSkill string,
	workStaffs []table.WorkStaff,
	staffSkills []table.StaffSkill,
	jst *time.Location,
) []string {
	eligible := make([]string, 0)
	seen := make(map[string]bool)

	for _, staff := range workStaffs {
		if staff.AliasName == "" || seen[staff.AliasName] {
			continue
		}
		workStart, startErr := receptionBookParseJST(staff.WorkStart, jst)
		workEnd, endErr := receptionBookParseJST(staff.WorkEnd, jst)
		if startErr != nil || endErr != nil || !workEnd.After(workStart) {
			continue
		}
		if start.Before(workStart) || end.After(workEnd) {
			continue
		}
		if needSkill != "" && !receptionBookStaffHasSkill(staff.AliasName, needSkill, staffSkills) {
			continue
		}
		seen[staff.AliasName] = true
		eligible = append(eligible, staff.AliasName)
	}

	sort.Strings(eligible)
	return eligible
}

func intervalsOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && aEnd.After(bStart)
}

func receptionBookOverlapComponent(
	appointments []receptionBookAppointment,
	targetIndex int,
) []receptionBookAppointment {
	if targetIndex < 0 || targetIndex >= len(appointments) {
		return nil
	}

	inComponent := make([]bool, len(appointments))
	inComponent[targetIndex] = true
	changed := true
	for changed {
		changed = false
		for i := range appointments {
			if inComponent[i] {
				continue
			}
			for j := range appointments {
				if !inComponent[j] {
					continue
				}
				if intervalsOverlap(
					appointments[i].Start, appointments[i].End,
					appointments[j].Start, appointments[j].End,
				) {
					inComponent[i] = true
					changed = true
					break
				}
			}
		}
	}

	component := make([]receptionBookAppointment, 0)
	for i, appointment := range appointments {
		if inComponent[i] {
			component = append(component, appointment)
		}
	}
	return component
}

// receptionBookCanAssignStaff checks whether every appointment in an overlapping
// component can be assigned to a qualified working staff member, without assigning
// the same person to two overlapping appointments. Assignments are recalculated
// from current data; no specific staff member is persisted on a booking.
func receptionBookCanAssignStaff(appointments []receptionBookAppointment) bool {
	if len(appointments) == 0 {
		return false
	}

	order := make([]int, len(appointments))
	for i := range appointments {
		if len(appointments[i].Candidates) == 0 {
			return false
		}
		order[i] = i
	}

	// Try appointments with the fewest eligible staff first to reduce backtracking.
	sort.SliceStable(order, func(i, j int) bool {
		return len(appointments[order[i]].Candidates) < len(appointments[order[j]].Candidates)
	})

	assigned := make(map[string][]int)
	var assign func(int) bool
	assign = func(position int) bool {
		if position == len(order) {
			return true
		}

		appointmentIndex := order[position]
		appointment := appointments[appointmentIndex]
		for _, aliasName := range appointment.Candidates {
			conflicts := false
			for _, assignedIndex := range assigned[aliasName] {
				other := appointments[assignedIndex]
				if intervalsOverlap(appointment.Start, appointment.End, other.Start, other.End) {
					conflicts = true
					break
				}
			}
			if conflicts {
				continue
			}

			assigned[aliasName] = append(assigned[aliasName], appointmentIndex)
			if assign(position + 1) {
				return true
			}
			assigned[aliasName] = assigned[aliasName][:len(assigned[aliasName])-1]
			if len(assigned[aliasName]) == 0 {
				delete(assigned, aliasName)
			}
		}
		return false
	}

	return assign(0)
}
