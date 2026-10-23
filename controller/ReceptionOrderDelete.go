package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/common"
	"chat/table"
)

func ReceptionOrderDelete(w http.ResponseWriter, r *http.Request) {
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID format", http.StatusOK)
		return
	}

	code := r.FormValue("code")
	aliasName := r.FormValue("aliasName")
	channelIDText := r.FormValue("channelID")
	facilityIDText := r.FormValue("facilityID")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	var reception table.ReceptionStruct
	var facilitiesJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, order_user_ids, facilities
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID,
		&reception.ChannelID,
		pq.Array(&reception.OrderUserIDs),
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
	if err := json.Unmarshal(facilitiesJSON, &reception.Facilities); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to decode reception facilities: "+err.Error(), http.StatusOK)
		return
	}

	// Verify staff against PostgreSQL. Inquiry-only aliases (access_right = 3)
	// are not accepted as shop staff.
	staffAccess := false
	if aliasName != "" && channelIDText != "" {
		channelID, parseErr := strconv.ParseInt(channelIDText, 10, 64)
		if parseErr == nil && channelID == reception.ChannelID {
			err = common.DB.AliasDB.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM alias
					WHERE user_id = $1
					  AND alias_name = $2
					  AND channel_id = $3
					  AND COALESCE(access_right, 0) IN (0, 1)
				)
			`, session.UserID, aliasName, channelID).Scan(&staffAccess)
			if err != nil {
				common.WriteResponseWithSession(w, session, "Failed to validate staff access: "+err.Error(), http.StatusOK)
				return
			}
		}
	}

	// Resolve the target facility from the separate current-code registry or
	// from reception.facilities JSONB. The JSONB master does not contain currentCode.
	var facility table.FacilityMaster
	foundByCode := false
	if code != "" {
		var facilityID int
		err = tx.QueryRowContext(ctx, `
			SELECT facility_id
			FROM reception_facility_code
			WHERE reception_id = $1 AND current_code = $2
		`, receptionID, code).Scan(&facilityID)
		switch {
		case err == nil:
			facility, foundByCode = receptionOrderDeleteFindFacility(reception.Facilities, facilityID)
			if !foundByCode {
				common.WriteResponseWithSession(w, session, "Facility code points to a facility missing from reception master data", http.StatusOK)
				return
			}
		case err == sql.ErrNoRows:
			// Verified staff may explicitly select a facility below.
		default:
			common.WriteResponseWithSession(w, session, "Failed to find facility code: "+err.Error(), http.StatusOK)
			return
		}
	}

	if !foundByCode {
		if !staffAccess {
			if code == "" {
				common.WriteResponseWithSession(w, session, "code is required unless you are verified staff", http.StatusOK)
			} else {
				common.WriteResponseWithSession(w, session, "コードが一致していません。スタッフでもありません", http.StatusOK)
			}
			return
		}

		facilityID, parseErr := strconv.Atoi(facilityIDText)
		if parseErr != nil || facilityID <= 0 {
			common.WriteResponseWithSession(w, session, "A valid facilityID is required for staff when code is not supplied or does not match", http.StatusOK)
			return
		}
		var found bool
		facility, found = receptionOrderDeleteFindFacility(reception.Facilities, facilityID)
		if !found {
			common.WriteResponseWithSession(w, session, "Facility not found", http.StatusOK)
			return
		}
	}

	// Load push destinations before mutating the registry so a read failure does
	// not leave the facility code cleared without notifying clients.
	userIDs := receptionOrderDeleteUniquePositiveUserIDs(reception.OrderUserIDs)
	sessions, err := receptionOrderDeleteLoadSessions(ctx, common.DB.SessionDB, userIDs)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load notification sessions: "+err.Error(), http.StatusOK)
		return
	}

	var result sql.Result
	if foundByCode {
		result, err = tx.ExecContext(ctx, `
			DELETE FROM reception_facility_code
			WHERE reception_id = $1 AND facility_id = $2 AND current_code = $3
		`, receptionID, facility.FacilityID, code)
	} else {
		// Verified staff selected a facility explicitly. Delete only its active code.
		result, err = tx.ExecContext(ctx, `
			DELETE FROM reception_facility_code
			WHERE reception_id = $1 AND facility_id = $2
		`, receptionID, facility.FacilityID)
	}
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to clear facility code: "+err.Error(), http.StatusOK)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to verify facility update: "+err.Error(), http.StatusOK)
		return
	}
	if rowsAffected != 1 {
		common.WriteResponseWithSession(w, session, "Facility code was already changed or is not set; please retry", http.StatusOK)
		return
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit facility update: "+err.Error(), http.StatusOK)
		return
	}

	// Preserve the existing push payload. Use the verified alias for staff and
	// the authenticated numeric user ID for customers.
	actorID := strconv.FormatInt(session.UserID, 10)
	if staffAccess {
		actorID = aliasName
	}
	payload := []interface{}{
		"receptionOrder",
		reception.ChannelID,
		actorID,
		facility.FacilityName,
	}
	common.ChunkPush(sessions, payload)

	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func receptionOrderDeleteFindFacility(facilities []table.FacilityMaster, facilityID int) (table.FacilityMaster, bool) {
	for _, facility := range facilities {
		if facility.FacilityID == facilityID {
			return facility, true
		}
	}
	return table.FacilityMaster{}, false
}

func receptionOrderDeleteLoadSessions(
	ctx context.Context,
	db *sql.DB,
	userIDs []int64,
) ([]table.SessionStruct, error) {
	if len(userIDs) == 0 {
		return []table.SessionStruct{}, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT session_id,
		       COALESCE(device_id, ''),
		       user_id,
		       COALESCE(created_at, CURRENT_TIMESTAMP),
		       COALESCE(updated_at, CURRENT_TIMESTAMP),
		       COALESCE(push_token, ''),
		       COALESCE(device_type, 0),
		       COALESCE(csrf, ''),
		       COALESCE(push_contents, ARRAY[]::TEXT[]),
		       COALESCE(mail, ''),
		       COALESCE(telephone, '')
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
		var item table.SessionStruct
		if err := rows.Scan(
			&item.SessionID,
			&item.DeviceID,
			&item.UserID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.PushToken,
			&item.DeviceType,
			&item.Csrf,
			pq.Array(&item.PushContents),
			&item.Mail,
			&item.Telephone,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, item)
	}
	return sessions, rows.Err()
}

func receptionOrderDeleteUniquePositiveUserIDs(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
