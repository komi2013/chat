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

// ReceptionOrder validates a customer order against the reception's JSONB
// menu master and sends a push notification to the configured recipients.
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

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	// Reception/menu master data is stored in reception arrays + JSONB columns.
	var reception table.ReceptionStruct
	var menusJSON, itemDetailsJSON, facilitiesJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, order_user_ids,
		       menus, item_details, facilities
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID,
		&reception.ChannelID,
		pq.Array(&reception.OrderUserIDs),
		&menusJSON,
		&itemDetailsJSON,
		&facilitiesJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		}
		return
	}

	if err := json.Unmarshal(menusJSON, &reception.Menus); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to decode reception menus: "+err.Error(), http.StatusOK)
		return
	}
	if err := json.Unmarshal(itemDetailsJSON, &reception.ItemDetails); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to decode reception item details: "+err.Error(), http.StatusOK)
		return
	}
	if err := json.Unmarshal(facilitiesJSON, &reception.Facilities); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to decode reception facilities: "+err.Error(), http.StatusOK)
		return
	}

	// Active facility codes are kept in the operational registry, not in the
	// facilities JSONB array. Resolve the stable facility ID, then find its master.
	var facilityID int
	err = tx.QueryRowContext(ctx, `
		SELECT facility_id
		FROM reception_facility_code
		WHERE reception_id = $1 AND current_code = $2
	`, receptionID, code).Scan(&facilityID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "コードが一致してません", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to validate facility code: "+err.Error(), http.StatusOK)
		}
		return
	}

	var facility table.FacilityMaster
	facilityFound := false
	for _, item := range reception.Facilities {
		if item.FacilityID == facilityID {
			facility = item
			facilityFound = true
			break
		}
	}
	if !facilityFound {
		common.WriteResponseWithSession(w, session, "Facility not found", http.StatusOK)
		return
	}

	menus := make(map[int64]table.Menu, len(reception.Menus))
	for _, menu := range reception.Menus {
		// Duplicate IDs would make order validation ambiguous; reject malformed
		// master data rather than silently choosing one of the duplicate entries.
		if _, exists := menus[menu.MenuID]; exists {
			common.WriteResponseWithSession(w, session, fmt.Sprintf("Duplicate menuID %d in reception configuration", menu.MenuID), http.StatusOK)
			return
		}
		menus[menu.MenuID] = menu
	}

	configs := make(map[int64]*receptionOrderMenuConfig, len(menus))
	for _, menu := range menus {
		configs[menu.MenuID] = receptionOrderBuildMenuConfig(menu, reception.ItemDetails)
	}

	// Validate every order before notifying anyone. Menu/options and prices are
	// taken from the database-backed JSONB master, never trusted from the client.
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

	// Keep the existing array payload shape. Index 2 contains the numeric
	// facility ID rather than a personal nickname.
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

// receptionOrderBuildMenuConfig derives option/choice validation data from the
// JSONB-backed table.Menu and table.ItemDetail structs.
func receptionOrderBuildMenuConfig(menu table.Menu, details []table.ItemDetail) *receptionOrderMenuConfig {
	cfg := &receptionOrderMenuConfig{
		itemIDs:      make(map[int64]bool, len(menu.Items)),
		paidPrices:   make(map[int64]int, len(menu.PaidOptions)),
		freeGroups:   menu.FreeOptions,
		freeMultiIDs: make(map[int64]bool, len(menu.FreeMultiOptions)),
		itemChoices:  make(map[int64]map[string]bool),
	}
	for _, id := range menu.Items {
		cfg.itemIDs[id] = true
	}
	for _, option := range menu.PaidOptions {
		cfg.paidPrices[option.ItemID] = option.Price
	}
	for _, id := range menu.FreeMultiOptions {
		cfg.freeMultiIDs[id] = true
	}
	for _, detail := range details {
		if !cfg.itemIDs[detail.ItemID] {
			continue
		}
		if cfg.itemChoices[detail.ItemID] == nil {
			cfg.itemChoices[detail.ItemID] = make(map[string]bool)
		}
		for _, group := range detail.Choices {
			for _, choice := range group {
				cfg.itemChoices[detail.ItemID][choice] = true
			}
		}
	}
	return cfg
}

func receptionOrderValidate(order *ReceptionOrderStruct, menu table.Menu, cfg *receptionOrderMenuConfig) error {
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

	// Normalize display and price fields from database-verified values.
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

	// Match the most constrained selection first so every selection maps to a
	// distinct free-option group, even when item IDs appear in multiple groups.
	sort.SliceStable(candidates, func(i, j int) bool {
		return len(candidates[i].groups) < len(candidates[j].groups)
	})
	used := make([]bool, len(groups))
	var match func(int) bool
	match = func(pos int) bool {
		if pos == len(candidates) {
			return true
		}
		for _, groupIndex := range candidates[pos].groups {
			if used[groupIndex] {
				continue
			}
			used[groupIndex] = true
			if match(pos + 1) {
				return true
			}
			used[groupIndex] = false
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
		FROM "session"
		WHERE user_id = ANY($1::BIGINT[])
		ORDER BY user_id, created_at
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]table.SessionStruct, 0)
	for rows.Next() {
		var s table.SessionStruct
		if err := rows.Scan(
			&s.SessionID, &s.DeviceID, &s.UserID, &s.CreatedAt, &s.UpdatedAt,
			&s.PushToken, &s.DeviceType, &s.Csrf, pq.Array(&s.PushContents), &s.Mail, &s.Telephone,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
