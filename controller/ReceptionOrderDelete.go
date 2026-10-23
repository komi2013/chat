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
	table "chat/database"
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
	err = tx.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, order_user_ids
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID,
		&reception.ChannelID,
		pq.Array(&reception.OrderUserIDs),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		}
		return
	}

	// Staff authorization is verified against PostgreSQL rather than session aliases.
	// Inquirer aliases (access_right = 3) are not accepted as shop staff.
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

	// A customer can authorize with the current facility code. Verified staff can
	// authorize without a code by specifying facilityID.
	var facility table.ReceptionFacilityStruct
	foundByCode := false
	if code != "" {
		err = tx.QueryRowContext(ctx, `
			SELECT facility_id, reception_id, sort_order, facility_name,
			       COALESCE(capacity, 0), COALESCE(current_code, ''), bookable
			FROM reception_facility
			WHERE reception_id = $1 AND current_code = $2
			ORDER BY sort_order, facility_id
			LIMIT 1
		`, receptionID, code).Scan(
			&facility.FacilityID,
			&facility.ReceptionID,
			&facility.SortOrder,
			&facility.FacilityName,
			&facility.Capacity,
			&facility.CurrentCode,
			&facility.Bookable,
		)
		switch {
		case err == nil:
			foundByCode = true
		case err == sql.ErrNoRows:
			// Staff may fall back to facilityID.
		default:
			common.WriteResponseWithSession(w, session, "Failed to find facility: "+err.Error(), http.StatusOK)
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

		facilityID, parseErr := strconv.ParseInt(facilityIDText, 10, 64)
		if parseErr != nil || facilityID <= 0 {
			common.WriteResponseWithSession(w, session, "A valid facilityID is required for staff when code is not supplied or does not match", http.StatusOK)
			return
		}

		err = tx.QueryRowContext(ctx, `
			SELECT facility_id, reception_id, sort_order, facility_name,
			       COALESCE(capacity, 0), COALESCE(current_code, ''), bookable
			FROM reception_facility
			WHERE reception_id = $1 AND facility_id = $2
		`, receptionID, facilityID).Scan(
			&facility.FacilityID,
			&facility.ReceptionID,
			&facility.SortOrder,
			&facility.FacilityName,
			&facility.Capacity,
			&facility.CurrentCode,
			&facility.Bookable,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				common.WriteResponseWithSession(w, session, "Facility not found", http.StatusOK)
			} else {
				common.WriteResponseWithSession(w, session, "Failed to find facility: "+err.Error(), http.StatusOK)
			}
			return
		}
	}

	// Read notification destinations before clearing the code. If this read fails,
	// the DB change is rolled back and the caller can retry.
	userIDs := receptionOrderDeleteUniquePositiveUserIDs(reception.OrderUserIDs)
	sessions, err := receptionOrderDeleteLoadSessions(ctx, common.DB.SessionDB, userIDs)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load notification sessions: "+err.Error(), http.StatusOK)
		return
	}

	var result sql.Result
	if foundByCode {
		result, err = tx.ExecContext(ctx, `
			UPDATE reception_facility
			SET current_code = ''
			WHERE reception_id = $1
			  AND facility_id = $2
			  AND current_code = $3
		`, receptionID, facility.FacilityID, code)
	} else {
		// Verified staff selected a facility explicitly.
		result, err = tx.ExecContext(ctx, `
			UPDATE reception_facility
			SET current_code = ''
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
		common.WriteResponseWithSession(w, session, "Facility code was already changed; please retry", http.StatusOK)
		return
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit facility update: "+err.Error(), http.StatusOK)
		return
	}

	// Keep the old push message shape. Staff uses the verified alias; a customer
	// uses their numeric user ID instead of an unverified alias from the request.
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
