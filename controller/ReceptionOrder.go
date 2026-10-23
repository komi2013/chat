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

func ReceptionOrder(w http.ResponseWriter, r *http.Request) {
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID format", http.StatusOK)
		return
	}
	code := r.FormValue("code")
	if code == "" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "code is invalid", http.StatusOK)
		return
	}

	var orders []ReceptionOrderStruct
	if err := json.Unmarshal([]byte(r.FormValue("receptionOrders")), &orders); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "JSON receptionOrders is wrong", http.StatusOK)
		return
	}
	if len(orders) == 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "No reception orders", http.StatusOK)
		return
	}
	for _, order := range orders {
		if order.ReceptionOrderID == "" || order.SeatName == "" || order.MenuID <= 0 {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "data is not enough", http.StatusOK)
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

	// Reception/menu configuration is read from PostgreSQL table-by-table.
	tx, err := common.DB.ReceptionDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	var reception table.ReceptionStruct
	err = tx.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, admin_names, join_names, inquiry_names,
		       subscriptions, order_user_ids, COALESCE(updated_at, CURRENT_TIMESTAMP),
		       COALESCE(reception_title, ''), skills, work_staff_need
		FROM reception WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID, &reception.ChannelID,
		pq.Array(&reception.AdminNames), pq.Array(&reception.JoinNames),
		pq.Array(&reception.InquiryNames), pq.Array(&reception.Subscriptions),
		pq.Array(&reception.OrderUserIDs), &reception.UpdatedAt,
		&reception.ReceptionTitle, pq.Array(&reception.Skills), &reception.WorkStaffNeed,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		}
		return
	}

	// The valid facility code supplies both the visible seat name and the
	// non-personal identifier inserted in the notification payload.
	var facility table.ReceptionFacilityStruct
	err = tx.QueryRowContext(ctx, `
		SELECT facility_id, reception_id, sort_order, facility_name,
		       COALESCE(capacity, 0), COALESCE(current_code, ''), bookable
		FROM reception_facility
		WHERE reception_id = $1 AND current_code = $2
		ORDER BY sort_order, facility_id LIMIT 1
	`, receptionID, code).Scan(
		&facility.FacilityID, &facility.ReceptionID, &facility.SortOrder,
		&facility.FacilityName, &facility.Capacity, &facility.CurrentCode, &facility.Bookable,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "コードが一致してません", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to validate facility code: "+err.Error(), http.StatusOK)
		}
		return
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, reception_id, sort_order, menu_id, COALESCE(menu_name, ''),
		       COALESCE(price, 0), COALESCE(prepaid_price, 0),
		       COALESCE(need_skill, ''), COALESCE(need_facility, ''),
		       COALESCE(specify_name_flag, FALSE), COALESCE(spend_minute, 0),
		       COALESCE(for_book_type, 0)
		FROM reception_menu_item WHERE reception_id = $1 ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load menu: "+err.Error(), http.StatusOK)
		return
	}
	menus := make(map[int64]table.ReceptionMenuItemStruct)
	for rows.Next() {
		var m table.ReceptionMenuItemStruct
		if err := rows.Scan(&m.ID, &m.ReceptionID, &m.SortOrder, &m.MenuID, &m.MenuName,
			&m.Price, &m.PrepaidPrice, &m.NeedSkill, &m.NeedFacility,
			&m.SpecifyNameFlag, &m.SpendMinute, &m.ForBookType); err != nil {
			rows.Close()
			common.WriteResponseWithSession(w, session, "Failed to read menu: "+err.Error(), http.StatusOK)
			return
		}
		menus[m.MenuID] = m
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		common.WriteResponseWithSession(w, session, "Failed to read menu: "+err.Error(), http.StatusOK)
		return
	}
	rows.Close()

	configs := make(map[int64]*receptionOrderMenuConfig, len(menus))
	for _, menu := range menus {
		cfg, err := receptionOrderLoadMenuConfig(ctx, tx, receptionID, menu)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to load menu options: "+err.Error(), http.StatusOK)
			return
		}
		configs[menu.MenuID] = cfg
	}

	// Validate the entire payload before notifying anyone. Invalid menu IDs,
	// option IDs, and client-supplied prices reject the entire request.
	for i := range orders {
		menu, ok := menus[orders[i].MenuID]
		if !ok {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("MenuID %d not found", orders[i].MenuID), http.StatusOK)
			return
		}
		if err := receptionOrderValidate(&orders[i], menu, configs[menu.MenuID]); err != nil {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("Invalid order %s: %v", orders[i].ReceptionOrderID, err), http.StatusOK)
			return
		}
		orders[i].SeatName = facility.FacilityName
		orders[i].MenuName = menu.MenuName
	}
	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to finish validation: "+err.Error(), http.StatusOK)
		return
	}

	// Notify sessions for reception.order_user_ids plus the current user.
	// Deduplicate user IDs; ANY() also avoids duplicate rows for each session.
	userIDs := make([]int64, 0, len(reception.OrderUserIDs)+1)
	seen := make(map[int64]bool)
	for _, id := range reception.OrderUserIDs {
		if id > 0 && !seen[id] {
			seen[id] = true
			userIDs = append(userIDs, id)
		}
	}
	if session.UserID > 0 && !seen[session.UserID] {
		userIDs = append(userIDs, session.UserID)
	}
	sessions, err := receptionOrderLoadSessions(ctx, common.DB.SessionDB, userIDs)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load notification sessions: "+err.Error(), http.StatusOK)
		return
	}

	// Payload stays array-shaped for existing consumers. Index 2 is now the
	// numeric FacilityID instead of session.Nickname.
	payload := []interface{}{"receptionOrder", reception.ChannelID, facility.FacilityID, orders}
	common.ChunkPush(sessions, payload)

	response := common.BaseResponse{Csrf: session.Csrf, PushContents: session.PushContents}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

type ReceptionOrderStruct struct {
	ReceptionOrderID string            `json:"receptionOrderID"`
	SeatName         string            `json:"seatName"`
	MenuID           int64             `json:"menuID"`
	MenuName         string            `json:"menuName"`
	ItemChoices      map[string]string `json:"itemChoices"`
	FreeOptions      []int64           `json:"freeOptions"`
	PaidOptions      []struct {
		ItemID int64 `json:"itemID"`
		Price  int   `json:"price"`
	} `json:"paidOptions"`
	FreeMultiOptions []int64 `json:"freeMultiOptions"`
	Price            int     `json:"price"`
	TotalPrice       int     `json:"totalPrice"`
}

type receptionOrderMenuConfig struct {
	itemIDs      map[int64]bool
	paidPrices   map[int64]int
	freeGroups   [][]int64
	freeMultiIDs map[int64]bool
	itemChoices  map[int64]map[string]bool
}

func receptionOrderLoadMenuConfig(ctx context.Context, tx *sql.Tx, receptionID int64, menu table.ReceptionMenuItemStruct) (*receptionOrderMenuConfig, error) {
	cfg := &receptionOrderMenuConfig{
		itemIDs: make(map[int64]bool), paidPrices: make(map[int64]int),
		freeGroups: make([][]int64, 0), freeMultiIDs: make(map[int64]bool),
		itemChoices: make(map[int64]map[string]bool),
	}

	rows, err := tx.QueryContext(ctx, `SELECT item_id FROM reception_menu_item_content WHERE menu_item_id=$1 ORDER BY sort_order, id`, menu.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		cfg.itemIDs[id] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	rows, err = tx.QueryContext(ctx, `SELECT item_id, price FROM reception_menu_paid_option WHERE menu_item_id=$1 ORDER BY sort_order, id`, menu.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var price int
		if err := rows.Scan(&id, &price); err != nil {
			rows.Close()
			return nil, err
		}
		cfg.paidPrices[id] = price
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	rows, err = tx.QueryContext(ctx, `SELECT id FROM reception_menu_free_option_group WHERE menu_item_id=$1 ORDER BY sort_order, id`, menu.ID)
	if err != nil {
		return nil, err
	}
	groupIDs := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		groupIDs = append(groupIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for _, groupID := range groupIDs {
		items, err := tx.QueryContext(ctx, `SELECT item_id FROM reception_menu_free_option_item WHERE free_option_group_id=$1 ORDER BY sort_order, id`, groupID)
		if err != nil {
			return nil, err
		}
		group := make([]int64, 0)
		for items.Next() {
			var id int64
			if err := items.Scan(&id); err != nil {
				items.Close()
				return nil, err
			}
			group = append(group, id)
		}
		if err := items.Err(); err != nil {
			items.Close()
			return nil, err
		}
		items.Close()
		cfg.freeGroups = append(cfg.freeGroups, group)
	}

	rows, err = tx.QueryContext(ctx, `SELECT item_id FROM reception_menu_free_multi_option WHERE menu_item_id=$1 ORDER BY sort_order, id`, menu.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		cfg.freeMultiIDs[id] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	// Item choices are tied to an item in this reception; the item must also
	// belong to the selected menu (checked below).
	rows, err = tx.QueryContext(ctx, `
		SELECT d.item_id, c.choice_text
		FROM reception_menu_item_detail d
		JOIN reception_menu_item_detail_choice_group g ON g.item_detail_id=d.id
		JOIN reception_menu_item_detail_choice c ON c.choice_group_id=g.id
		WHERE d.reception_id=$1
		ORDER BY d.sort_order, g.sort_order, c.sort_order, c.id
	`, receptionID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var value string
		if err := rows.Scan(&id, &value); err != nil {
			rows.Close()
			return nil, err
		}
		if !cfg.itemIDs[id] {
			continue
		}
		if cfg.itemChoices[id] == nil {
			cfg.itemChoices[id] = make(map[string]bool)
		}
		cfg.itemChoices[id][value] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	return cfg, nil
}

func receptionOrderValidate(order *ReceptionOrderStruct, menu table.ReceptionMenuItemStruct, cfg *receptionOrderMenuConfig) error {
	if order.Price != menu.Price {
		return fmt.Errorf("price does not match the configured menu price")
	}
	paidSum := 0
	seenPaid := make(map[int64]bool)
	for _, option := range order.PaidOptions {
		price, ok := cfg.paidPrices[option.ItemID]
		if !ok {
			return fmt.Errorf("paid option item %d is not available", option.ItemID)
		}
		if seenPaid[option.ItemID] {
			return fmt.Errorf("paid option item %d is duplicated", option.ItemID)
		}
		if option.Price != price {
			return fmt.Errorf("paid option price mismatch for item %d", option.ItemID)
		}
		seenPaid[option.ItemID] = true
		paidSum += price
	}
	if order.TotalPrice != menu.Price+paidSum {
		return fmt.Errorf("totalPrice does not match menu price plus paid options")
	}
	if err := receptionOrderValidateFreeOptions(order.FreeOptions, cfg.freeGroups); err != nil {
		return err
	}
	seenMulti := make(map[int64]bool)
	for _, id := range order.FreeMultiOptions {
		if !cfg.freeMultiIDs[id] {
			return fmt.Errorf("free multi option item %d is not available", id)
		}
		if seenMulti[id] {
			return fmt.Errorf("free multi option item %d is duplicated", id)
		}
		seenMulti[id] = true
	}
	for key, value := range order.ItemChoices {
		id, err := strconv.ParseInt(key, 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid itemChoices key %q", key)
		}
		if !cfg.itemIDs[id] {
			return fmt.Errorf("item %d is not included in this menu", id)
		}
		if !cfg.itemChoices[id][value] {
			return fmt.Errorf("choice %q is not configured for item %d", value, id)
		}
	}
	// Normalize display and price fields from DB-verified values.
	order.MenuName = menu.MenuName
	order.Price = menu.Price
	order.TotalPrice = menu.Price + paidSum
	return nil
}

func receptionOrderValidateFreeOptions(selected []int64, groups [][]int64) error {
	if len(selected) == 0 {
		return nil
	}
	if len(selected) > len(groups) {
		return fmt.Errorf("too many free option selections")
	}
	type candidate struct{ groups []int }
	candidates := make([]candidate, 0, len(selected))
	seen := make(map[int64]bool)
	for _, selectedID := range selected {
		if seen[selectedID] {
			return fmt.Errorf("free option item %d is duplicated", selectedID)
		}
		seen[selectedID] = true
		c := candidate{}
		for gi, items := range groups {
			for _, itemID := range items {
				if selectedID == itemID {
					c.groups = append(c.groups, gi)
					break
				}
			}
		}
		if len(c.groups) == 0 {
			return fmt.Errorf("free option item %d is not available", selectedID)
		}
		candidates = append(candidates, c)
	}
	// Most constrained choice first, then find distinct groups for all selections.
	sort.SliceStable(candidates, func(i, j int) bool { return len(candidates[i].groups) < len(candidates[j].groups) })
	used := make([]bool, len(groups))
	var match func(int) bool
	match = func(pos int) bool {
		if pos == len(candidates) {
			return true
		}
		for _, gi := range candidates[pos].groups {
			if used[gi] {
				continue
			}
			used[gi] = true
			if match(pos + 1) {
				return true
			}
			used[gi] = false
		}
		return false
	}
	if !match(0) {
		return fmt.Errorf("free option selections do not match distinct groups")
	}
	return nil
}

func receptionOrderLoadSessions(ctx context.Context, db *sql.DB, userIDs []int64) ([]table.SessionStruct, error) {
	if len(userIDs) == 0 {
		return []table.SessionStruct{}, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT session_id, COALESCE(device_id, ''), user_id,
		       COALESCE(created_at, CURRENT_TIMESTAMP), COALESCE(updated_at, CURRENT_TIMESTAMP),
		       COALESCE(push_token, ''), COALESCE(device_type, 0), COALESCE(csrf, ''),
	       push_contents, COALESCE(mail, ''), COALESCE(telephone, '')
		FROM "session" WHERE user_id = ANY($1::BIGINT[]) ORDER BY user_id, created_at
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]table.SessionStruct, 0)
	for rows.Next() {
		var s table.SessionStruct
		if err := rows.Scan(&s.SessionID, &s.DeviceID, &s.UserID, &s.CreatedAt, &s.UpdatedAt,
			&s.PushToken, &s.DeviceType, &s.Csrf, pq.Array(&s.PushContents), &s.Mail, &s.Telephone); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
