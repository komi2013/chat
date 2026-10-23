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

func ReceptionEdit(w http.ResponseWriter, r *http.Request) {
	var reception collection.ReceptionStruct
	if err := json.Unmarshal([]byte(r.FormValue("reception")), &reception); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON reception", http.StatusOK)
		return
	}

	var menu collection.MenuStruct
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

	// Confirm the POSTed aliasName/channelID belongs to this logged-in user.
	// This check validates request consistency; edit authorization is checked separately
	// against reception.admin_names below.
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

	// Do not silently accept conflicting channel IDs from the JSON and the form.
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
		// A new reception has no stored admin_names yet. Require the creator to be
		// included in the initial admin_names list.
		if !containsString(reception.AdminNames, aliasName) {
			common.WriteResponseWithSession(w, session, "Only reception admins can create or edit this reception", http.StatusOK)
			return
		}
	} else {
		// Lock the current reception row and authorize against the stored admin_names,
		// never against admin_names supplied in the POST body.
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

	// Save item images after identity and authorization checks.
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

	now := time.Now()
	reception.UpdatedAt = now
	menu.UpdatedAt = now
	reception.ChannelID = channelID

	if isNewReception {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO reception (
				channel_id, admin_names, join_names, inquiry_names, subscriptions,
				order_user_ids, updated_at, reception_title, skills, work_staff_need
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING reception_id
		`,
			reception.ChannelID,
			pq.Array(nonNilStrings(reception.AdminNames)),
			pq.Array(nonNilStrings(reception.JoinNames)),
			pq.Array(nonNilStrings(reception.InquiryNames)),
			pq.Array(nonNilStrings(reception.Subscriptions)),
			pq.Array(nonNilInt64s(reception.OrderUserIDs)),
			reception.UpdatedAt,
			reception.ReceptionTitle,
			pq.Array(nonNilStrings(reception.Skills)),
			reception.WorkStaffNeed,
		).Scan(&reception.ReceptionID)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to insert reception: "+err.Error(), http.StatusOK)
			return
		}
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE reception
			SET channel_id = $1,
			    admin_names = $2,
			    join_names = $3,
			    inquiry_names = $4,
			    subscriptions = $5,
			    order_user_ids = $6,
			    updated_at = $7,
			    reception_title = $8,
			    skills = $9,
			    work_staff_need = $10
			WHERE reception_id = $11
		`,
			reception.ChannelID,
			pq.Array(nonNilStrings(reception.AdminNames)),
			pq.Array(nonNilStrings(reception.JoinNames)),
			pq.Array(nonNilStrings(reception.InquiryNames)),
			pq.Array(nonNilStrings(reception.Subscriptions)),
			pq.Array(nonNilInt64s(reception.OrderUserIDs)),
			reception.UpdatedAt,
			reception.ReceptionTitle,
			pq.Array(nonNilStrings(reception.Skills)),
			reception.WorkStaffNeed,
			reception.ReceptionID,
		)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to update reception: "+err.Error(), http.StatusOK)
			return
		}
	}

	if err = replaceReceptionChildren(ctx, tx, reception); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to save reception details: "+err.Error(), http.StatusOK)
		return
	}

	menu.ReceptionID = reception.ReceptionID
	if err = replaceReceptionMenu(ctx, tx, menu); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to save menu: "+err.Error(), http.StatusOK)
		return
	}

	if err = tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit reception transaction: "+err.Error(), http.StatusOK)
		return
	}

	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		// Response headers/body may already be written; no second response is sent.
		return
	}
}

func replaceReceptionChildren(ctx context.Context, tx *sql.Tx, reception collection.ReceptionStruct) error {
	id := reception.ReceptionID

	// Children with facility_id must be deleted before reception_facility rows.
	deletes := []string{
		`DELETE FROM reception_facility_passcode WHERE facility_id IN (SELECT facility_id FROM reception_facility WHERE reception_id = $1)`,
		`DELETE FROM reception_facility_book_time WHERE facility_id IN (SELECT facility_id FROM reception_facility WHERE reception_id = $1)`,
		`DELETE FROM reception_passcode WHERE reception_id = $1`,
		`DELETE FROM reception_open_time WHERE reception_id = $1`,
		`DELETE FROM reception_ask WHERE reception_id = $1`,
		`DELETE FROM reception_ask_choice WHERE reception_id = $1`,
		`DELETE FROM reception_ask_multi_choice WHERE reception_id = $1`,
		`DELETE FROM reception_facility WHERE reception_id = $1`,
		`DELETE FROM reception_shift WHERE reception_id = $1`,
		`DELETE FROM reception_staff_skill WHERE reception_id = $1`,
		`DELETE FROM reception_work_staff WHERE reception_id = $1`,
		`DELETE FROM reception_book WHERE reception_id = $1`,
		`DELETE FROM reception_queue WHERE reception_id = $1`,
		`DELETE FROM reception_wait_config WHERE reception_id = $1`,
	}
	for _, query := range deletes {
		if _, err := tx.ExecContext(ctx, query, id); err != nil {
			return err
		}
	}

	for i, p := range reception.Passcodes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_passcode (reception_id, sort_order, passkey, pass_start, pass_end)
			VALUES ($1, $2, $3, $4, $5)
		`, id, i, p.Passkey, p.PassStart, p.PassEnd); err != nil {
			return fmt.Errorf("insert passcode[%d]: %w", i, err)
		}
	}

	for i, t := range reception.OpenTimes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_open_time (reception_id, sort_order, limit_start, limit_end)
			VALUES ($1, $2, $3, $4)
		`, id, i, t.LimitStart, t.LimitEnd); err != nil {
			return fmt.Errorf("insert openTime[%d]: %w", i, err)
		}
	}

	for i, a := range reception.Asks {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_ask (reception_id, sort_order, question, sequence)
			VALUES ($1, $2, $3, $4)
		`, id, i, a.Question, a.Sequence); err != nil {
			return fmt.Errorf("insert ask[%d]: %w", i, err)
		}
	}

	for i, a := range reception.AskChoices {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_ask_choice (reception_id, sort_order, question, choices, sequence)
			VALUES ($1, $2, $3, $4, $5)
		`, id, i, a.Question, pq.Array(nonNilStrings(a.Choices)), a.Sequence); err != nil {
			return fmt.Errorf("insert askChoice[%d]: %w", i, err)
		}
	}

	for i, a := range reception.AskMultiChoices {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_ask_multi_choice (reception_id, sort_order, question, choices, sequence)
			VALUES ($1, $2, $3, $4, $5)
		`, id, i, a.Question, pq.Array(nonNilStrings(a.Choices)), a.Sequence); err != nil {
			return fmt.Errorf("insert askMultiChoice[%d]: %w", i, err)
		}
	}

	for i, f := range reception.Facilities {
		var facilityID int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO reception_facility (
				reception_id, sort_order, facility_name, capacity, current_code, bookable
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING facility_id
		`, id, i, f.FacilityName, f.Capacity, nullableString(f.CurrentCode), f.Bookable).Scan(&facilityID)
		if err != nil {
			return fmt.Errorf("insert facility[%d]: %w", i, err)
		}

		for j, p := range f.Passcodes {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO reception_facility_passcode (facility_id, sort_order, passkey, pass_start, pass_end)
				VALUES ($1, $2, $3, $4, $5)
			`, facilityID, j, p.Passkey, nullableString(p.PassStart), nullableString(p.PassEnd)); err != nil {
				return fmt.Errorf("insert facility[%d].passcodes[%d]: %w", i, j, err)
			}
		}
		for j, bt := range f.BookTimes {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO reception_facility_book_time (facility_id, sort_order, book_start, book_end)
				VALUES ($1, $2, $3, $4)
			`, facilityID, j, bt.BookStart, bt.BookEnd); err != nil {
				return fmt.Errorf("insert facility[%d].bookTimes[%d]: %w", i, j, err)
			}
		}
	}

	// Reinsert shifts while preserving existing shift IDs. New shifts (ShiftID == 0)
	// get an identity value from PostgreSQL and that value is copied back to the response.
	shiftIDs := make(map[int64]bool, len(reception.Shifts))
	for i, s := range reception.Shifts {
		var shiftID int64
		var err error
		if s.ShiftID > 0 {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO reception_shift (
					shift_id, reception_id, sort_order, alias_names, shift_start, shift_end, open, skill, fix
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING shift_id
			`, s.ShiftID, id, i, pq.Array(nonNilStrings(s.AliasNames)), s.ShiftStart, s.ShiftEnd, s.Open, s.Skill, s.Fix).Scan(&shiftID)
		} else {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO reception_shift (
					reception_id, sort_order, alias_names, shift_start, shift_end, open, skill, fix
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING shift_id
			`, id, i, pq.Array(nonNilStrings(s.AliasNames)), s.ShiftStart, s.ShiftEnd, s.Open, s.Skill, s.Fix).Scan(&shiftID)
		}
		if err != nil {
			return fmt.Errorf("insert shift[%d]: %w", i, err)
		}
		reception.Shifts[i].ShiftID = shiftID
		shiftIDs[shiftID] = true
	}

	for i, s := range reception.StaffSkills {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_staff_skill (reception_id, sort_order, alias_name, skills)
			VALUES ($1, $2, $3, $4)
		`, id, i, nullableString(s.AliasName), pq.Array(nonNilStrings(s.Skills))); err != nil {
			return fmt.Errorf("insert staffSkill[%d]: %w", i, err)
		}
	}

	for i, s := range reception.WorkStaffs {
		if s.ShiftID <= 0 || !shiftIDs[s.ShiftID] {
			return fmt.Errorf("insert workStaff[%d]: missing or invalid shiftID %d", i, s.ShiftID)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_work_staff (reception_id, shift_id, sort_order, alias_name, work_start, work_end, seq)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, s.ShiftID, i, s.AliasName, s.WorkStart, s.WorkEnd, s.Seq); err != nil {
			return fmt.Errorf("insert workStaff[%d]: %w", i, err)
		}
	}

	for i, b := range reception.Books {
		var menuID any
		if b.MenuID != 0 {
			menuID = b.MenuID
		}
		var err error
		if b.CreatedAt.IsZero() {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO reception_book (
					reception_id, sort_order, book_start, book_end, answers, menu_id, people, nickname
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, id, i, b.BookStart, b.BookEnd, pq.Array(nonNilStrings(b.Answers)), menuID, b.People, b.Nickname)
		} else {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO reception_book (
					reception_id, sort_order, book_start, book_end, answers, menu_id, people, created_at, nickname
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, id, i, b.BookStart, b.BookEnd, pq.Array(nonNilStrings(b.Answers)), menuID, b.People, b.CreatedAt, b.Nickname)
		}
		if err != nil {
			return fmt.Errorf("insert book[%d]: %w", i, err)
		}
	}

	for i, q := range reception.Queues {
		var userID any
		if q.UserID != 0 {
			userID = q.UserID
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_queue (reception_id, sort_order, waiting_guest, queue_name, queued_at, user_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, id, i, q.WaitingGuest, nullableString(q.QueueName), nullableString(q.QueuedAt), userID); err != nil {
			return fmt.Errorf("insert queue[%d]: %w", i, err)
		}
	}

	for i, wc := range reception.WaitConfigs {
		guestRange := []int{wc.GuestRange[0], wc.GuestRange[1]}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reception_wait_config (reception_id, sort_order, guest_range, wait_ratio)
			VALUES ($1, $2, $3, $4)
		`, id, i, pq.Array(guestRange), wc.WaitRatio); err != nil {
			return fmt.Errorf("insert waitConfig[%d]: %w", i, err)
		}
	}

	return nil
}

func replaceReceptionMenu(ctx context.Context, tx *sql.Tx, menu collection.MenuStruct) error {
	id := menu.ReceptionID

	// Remove dependent rows first, then replace all menu rows for this reception.
	deletes := []string{
		`DELETE FROM reception_menu_item_content WHERE menu_item_id IN (SELECT id FROM reception_menu_item WHERE reception_id = $1)`,
		`DELETE FROM reception_menu_paid_option WHERE menu_item_id IN (SELECT id FROM reception_menu_item WHERE reception_id = $1)`,
		`DELETE FROM reception_menu_free_option_item WHERE free_option_group_id IN (SELECT id FROM reception_menu_free_option_group WHERE menu_item_id IN (SELECT id FROM reception_menu_item WHERE reception_id = $1))`,
		`DELETE FROM reception_menu_free_option_group WHERE menu_item_id IN (SELECT id FROM reception_menu_item WHERE reception_id = $1)`,
		`DELETE FROM reception_menu_free_multi_option WHERE menu_item_id IN (SELECT id FROM reception_menu_item WHERE reception_id = $1)`,
		`DELETE FROM reception_menu_item WHERE reception_id = $1`,
		`DELETE FROM reception_menu_item_detail_choice WHERE choice_group_id IN (SELECT id FROM reception_menu_item_detail_choice_group WHERE item_detail_id IN (SELECT id FROM reception_menu_item_detail WHERE reception_id = $1))`,
		`DELETE FROM reception_menu_item_detail_choice_group WHERE item_detail_id IN (SELECT id FROM reception_menu_item_detail WHERE reception_id = $1)`,
		`DELETE FROM reception_menu_item_detail WHERE reception_id = $1`,
	}
	for _, query := range deletes {
		if _, err := tx.ExecContext(ctx, query, id); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO reception_menu (reception_id, updated_at)
		VALUES ($1, $2)
		ON CONFLICT (reception_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
	`, id, menu.UpdatedAt); err != nil {
		return fmt.Errorf("upsert menu header: %w", err)
	}

	for i, m := range menu.Menus {
		var menuItemID int64
		err := tx.QueryRowContext(ctx, `
			INSERT INTO reception_menu_item (
				reception_id, sort_order, menu_id, menu_name, price, prepaid_price,
				need_skill, need_facility, specify_name_flag, spend_minute, for_book_type
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`, id, i, m.MenuID, m.MenuName, m.Price, m.PrepaidPrice,
			nullableString(m.NeedSkill), nullableString(m.NeedFacility), m.SpecifyNameFlag,
			nullableInt(m.SpendMinute), nullableInt(m.ForBookType)).Scan(&menuItemID)
		if err != nil {
			return fmt.Errorf("insert menu[%d]: %w", i, err)
		}

		for j, itemID := range m.Items {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO reception_menu_item_content (menu_item_id, sort_order, item_id)
				VALUES ($1, $2, $3)
			`, menuItemID, j, itemID); err != nil {
				return fmt.Errorf("insert menu[%d].items[%d]: %w", i, j, err)
			}
		}
		for j, option := range m.PaidOptions {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO reception_menu_paid_option (menu_item_id, sort_order, item_id, price)
				VALUES ($1, $2, $3, $4)
			`, menuItemID, j, option.ItemID, option.Price); err != nil {
				return fmt.Errorf("insert menu[%d].paidOptions[%d]: %w", i, j, err)
			}
		}
		for j, group := range m.FreeOptions {
			var groupID int64
			if err := tx.QueryRowContext(ctx, `
				INSERT INTO reception_menu_free_option_group (menu_item_id, sort_order)
				VALUES ($1, $2) RETURNING id
			`, menuItemID, j).Scan(&groupID); err != nil {
				return fmt.Errorf("insert menu[%d].freeOptions group[%d]: %w", i, j, err)
			}
			for k, itemID := range group {
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO reception_menu_free_option_item (free_option_group_id, sort_order, item_id)
					VALUES ($1, $2, $3)
				`, groupID, k, itemID); err != nil {
					return fmt.Errorf("insert menu[%d].freeOptions[%d][%d]: %w", i, j, k, err)
				}
			}
		}
		for j, itemID := range m.FreeMultiOptions {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO reception_menu_free_multi_option (menu_item_id, sort_order, item_id)
				VALUES ($1, $2, $3)
			`, menuItemID, j, itemID); err != nil {
				return fmt.Errorf("insert menu[%d].freeMultiOptions[%d]: %w", i, j, err)
			}
		}
	}

	for i, detail := range menu.ItemDetails {
		var detailID int64
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO reception_menu_item_detail (reception_id, sort_order, item_id, item_name, img_path)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, id, i, detail.ItemID, detail.ItemName, nullableString(detail.ImgPath)).Scan(&detailID); err != nil {
			return fmt.Errorf("insert itemDetail[%d]: %w", i, err)
		}

		for j, choices := range detail.Choices {
			var groupID int64
			if err := tx.QueryRowContext(ctx, `
				INSERT INTO reception_menu_item_detail_choice_group (item_detail_id, sort_order)
				VALUES ($1, $2) RETURNING id
			`, detailID, j).Scan(&groupID); err != nil {
				return fmt.Errorf("insert itemDetail[%d].choices group[%d]: %w", i, j, err)
			}
			for k, choice := range choices {
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO reception_menu_item_detail_choice (choice_group_id, sort_order, choice_text)
					VALUES ($1, $2, $3)
				`, groupID, k, choice); err != nil {
					return fmt.Errorf("insert itemDetail[%d].choices[%d][%d]: %w", i, j, k, err)
				}
			}
		}
	}

	return nil
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

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
