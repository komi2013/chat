package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"chat/common"
	"chat/table"

	"github.com/lib/pq"
)

const (
	accessRightNormal   int16 = 0
	accessRightAdmin    int16 = 1
	accessRightGuest    int16 = 2
	accessRightInquirer int16 = 3
)

type joinAlias struct {
	AliasID     string
	ChannelID   int64
	AliasName   string
	AliasImg    sql.NullString
	UserID      sql.NullString
	AliasBio    sql.NullString
	AccessRight sql.NullInt64
}

type joinGroup struct {
	GroupName  string
	AliasNames []string
	GroupImg   sql.NullString
}

type pendingJoinPush struct {
	session table.SessionStruct
	payload []interface{}
	pushID  string
}

// ChannelJoin joins a channel using an invitation code.
func ChannelJoin(w http.ResponseWriter, r *http.Request) {
	channelIDText := r.FormValue("channelID")
	myname := r.FormValue("myname")
	myimg := r.FormValue("myimg")
	code := r.FormValue("code")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		writeChannelJoinError(w, http.StatusUnauthorized, err.Error(), "")
		return
	}

	channelID, err := strconv.ParseInt(channelIDText, 10, 64)
	if err != nil || channelID <= 0 {
		writeChannelJoinError(w, http.StatusBadRequest, "Invalid channelID", session.Csrf)
		return
	}

	userID := session.UserID

	if myname == "" {
		writeChannelJoinError(w, http.StatusBadRequest, "myname is required", session.Csrf)
		return
	}

	// 1. Validate the channel and invitation.
	var (
		channelName         string
		channelDescription  sql.NullString
		invitationCode      sql.NullString
		guestInvitationCode sql.NullString
		invitedAt           sql.NullTime
	)

	err = common.DB.ChannelDB.QueryRow(`
		SELECT
			channel_name,
			channel_description,
			invitation_code,
			invitation_guest_code,
			invited_at
		FROM channel
		WHERE channel_id = $1
	`, channelID).Scan(
		&channelName,
		&channelDescription,
		&invitationCode,
		&guestInvitationCode,
		&invitedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeChannelJoinError(w, http.StatusNotFound, "Channel not found", session.Csrf)
			return
		}
		log.Printf("ChannelJoin: load channel: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel", session.Csrf)
		return
	}

	isRegularInvitation := invitationCode.Valid && invitationCode.String == code
	isGuestInvitation := guestInvitationCode.Valid && guestInvitationCode.String == code

	if (!isRegularInvitation && !isGuestInvitation) ||
		!invitedAt.Valid ||
		time.Since(invitedAt.Time) > 7*24*time.Hour {
		writeChannelJoinError(w, http.StatusBadRequest, "Invalid or expired invitation code", session.Csrf)
		return
	}

	// 2. Check alias-name uniqueness within this channel.
	var duplicate bool
	err = common.DB.AliasDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE channel_id = $1
			  AND alias_name = $2
			  AND COALESCE(access_right, 0) <> $3
		)
	`, channelID, myname, accessRightInquirer).Scan(&duplicate)
	if err != nil {
		log.Printf("ChannelJoin: check duplicate alias name: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to check alias name", session.Csrf)
		return
	}
	if duplicate {
		writeChannelJoinError(w, http.StatusConflict, "This alias name is already used in this channel", session.Csrf)
		return
	}

	// 3. Determine access right from the invitation code.
	var accessRight int16 = accessRightNormal
	if isGuestInvitation {
		accessRight = accessRightGuest
	}

	// 4. Save the image. ImgSave expects a numeric channel ID.
	if _, err := common.ImgSave(myimg, session.UserID, myname, channelID, 3, 1); err != nil {
		log.Printf("ChannelJoin: save alias image: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to save alias image", session.Csrf)
		return
	}

	// 5. Insert the alias.
	aliasID := channelIDText + myname
	_, err = common.DB.AliasDB.Exec(`
		INSERT INTO alias (
			alias_id,
			channel_id,
			alias_name,
			alias_img,
			user_id,
			access_right
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, aliasID, channelID, myname, myimg, session.UserID, accessRight)
	if err != nil {
		log.Printf("ChannelJoin: insert alias: %v", err)
		writeChannelJoinError(w, http.StatusConflict, "Failed to join channel; the alias name may already be in use", session.Csrf)
		return
	}

	// 6. Load aliases and groups for the initial sync.
	aliases, err := loadJoinAliases(channelID)
	if err != nil {
		log.Printf("ChannelJoin: load aliases: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel aliases", session.Csrf)
		return
	}

	groups, err := loadJoinGroups(channelID)
	if err != nil {
		log.Printf("ChannelJoin: load groups: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel groups", session.Csrf)
		return
	}

	// 7. Collect existing member user IDs, excluding inquirers.
	rows, err := common.DB.AliasDB.Query(`
		SELECT DISTINCT user_id
		FROM alias
		WHERE channel_id = $1
		  AND COALESCE(access_right, 0) <> $2
		  AND user_id IS NOT NULL
	`, channelID, accessRightInquirer)
	if err != nil {
		log.Printf("ChannelJoin: load member user IDs: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel members", session.Csrf)
		return
	}

	userIDs := make([]int64, 0)
	seenUserIDs := make(map[int64]bool)

	for rows.Next() {
		var memberUserID int64
		if err := rows.Scan(&memberUserID); err != nil {
			_ = rows.Close()
			log.Printf("ChannelJoin: scan member user ID: %v", err)
			writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel members", session.Csrf)
			return
		}

		if !seenUserIDs[memberUserID] {
			seenUserIDs[memberUserID] = true
			userIDs = append(userIDs, memberUserID)
		}
	}

	if err := rows.Err(); err != nil {
		_ = rows.Close()
		log.Printf("ChannelJoin: iterate member user IDs: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load channel members", session.Csrf)
		return
	}
	_ = rows.Close()

	// 8. Load sessions for all channel members.
	sessionRows, err := common.DB.SessionDB.Query(`
		SELECT session_id, user_id, COALESCE(push_token, ''), COALESCE(device_type, 0)
		FROM "session"
		WHERE user_id = ANY($1)
	`, pq.Array(userIDs))
	if err != nil {
		log.Printf("ChannelJoin: load sessions: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load member sessions", session.Csrf)
		return
	}

	var ownSessions []table.SessionStruct
	var otherSessions []table.SessionStruct

	for sessionRows.Next() {
		var (
			sessionID  string
			memberID   int64
			pushToken  string
			deviceType int16
		)

		if err := sessionRows.Scan(&sessionID, &memberID, &pushToken, &deviceType); err != nil {
			_ = sessionRows.Close()
			log.Printf("ChannelJoin: scan session: %v", err)
			writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load member sessions", session.Csrf)
			return
		}

		memberSession := table.SessionStruct{
			SessionID:  sessionID,
			UserID:     memberID,
			PushToken:  pushToken,
			DeviceType: int(deviceType),
		}

		if memberID == userID {
			ownSessions = append(ownSessions, memberSession)
		} else {
			otherSessions = append(otherSessions, memberSession)
		}
	}

	if err := sessionRows.Err(); err != nil {
		_ = sessionRows.Close()
		log.Printf("ChannelJoin: iterate sessions: %v", err)
		writeChannelJoinError(w, http.StatusInternalServerError, "Failed to load member sessions", session.Csrf)
		return
	}
	_ = sessionRows.Close()

	// 9. Prepare push payloads.
	var pending []pendingJoinPush

	for _, ownSession := range ownSessions {
		pending = append(pending, pendingJoinPush{
			session: ownSession,
			pushID:  channelIDText,
			payload: []interface{}{
				"channelEdit",
				channelIDText,
				channelName,
				nullStringValue(channelDescription),
			},
		})

		for _, alias := range aliases {
			pending = append(pending, pendingJoinPush{
				session: ownSession,
				pushID:  channelIDText,
				payload: []interface{}{
					"alias",
					channelIDText,
					myname,
					[]string{
						nullStringValue(alias.UserID),
						alias.AliasName,
						nullStringValue(alias.AliasBio),
						strconv.FormatInt(alias.AccessRight.Int64, 10),
					},
					nullStringValue(alias.AliasImg),
				},
			})
		}

		for _, group := range groups {
			pending = append(pending, pendingJoinPush{
				session: ownSession,
				pushID:  channelIDText,
				payload: []interface{}{
					"group",
					channelIDText,
					myname,
					[]interface{}{
						group.GroupName,
						group.AliasNames,
					},
					nullStringValue(group.GroupImg),
				},
			})
		}
	}

	// Notify other members about the newly joined alias only.
	for _, otherSession := range otherSessions {
		pending = append(pending, pendingJoinPush{
			session: otherSession,
			pushID:  channelIDText,
			payload: []interface{}{
				"alias",
				channelIDText,
				myname,
				[]string{
					strconv.FormatInt(session.UserID, 10),
					myname,
					"",
					strconv.Itoa(int(accessRight)),
				},
				myimg,
			},
		})
	}

	// 10. Queue all push contents before dispatching notifications.
	for _, p := range pending {
		data, err := json.Marshal(p.payload)
		if err != nil {
			log.Printf("ChannelJoin: marshal push payload: %v", err)
			writeChannelJoinError(w, http.StatusInternalServerError, "Failed to prepare notifications", session.Csrf)
			return
		}

		_, err = common.DB.SessionDB.Exec(`
			UPDATE "session"
			SET push_contents = array_append(
				COALESCE(push_contents, ARRAY[]::TEXT[]),
				$1
			)
			WHERE session_id = $2
		`, string(data), p.session.SessionID)
		if err != nil {
			log.Printf("ChannelJoin: queue push content for session %s: %v", p.session.SessionID, err)
			writeChannelJoinError(w, http.StatusInternalServerError, "Failed to queue notifications", session.Csrf)
			return
		}
	}

	// 11. Dispatch push notifications after database writes succeed.
	for _, p := range pending {
		if _, err := common.SendWebJustPush(p.payload, p.pushID, p.session); err != nil {
			log.Printf("ChannelJoin: push delivery failed for session %s: %v", p.session.SessionID, err)
		}
	}

	// 12. Respond to the client.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"csrf":         session.Csrf,
		"pushContents": session.PushContents,
	})
}

func loadJoinAliases(channelID int64) ([]joinAlias, error) {
	rows, err := common.DB.AliasDB.Query(`
		SELECT
			alias_id,
			channel_id,
			alias_name,
			alias_img,
			user_id,
			alias_bio,
			access_right
		FROM alias
		WHERE channel_id = $1
		ORDER BY alias_name
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aliases := make([]joinAlias, 0)
	for rows.Next() {
		var a joinAlias
		if err := rows.Scan(
			&a.AliasID,
			&a.ChannelID,
			&a.AliasName,
			&a.AliasImg,
			&a.UserID,
			&a.AliasBio,
			&a.AccessRight,
		); err != nil {
			return nil, err
		}
		aliases = append(aliases, a)
	}

	return aliases, rows.Err()
}

func loadJoinGroups(channelID int64) ([]joinGroup, error) {
	rows, err := common.DB.GroupDB.Query(`
		SELECT group_name, alias_names, group_img
		FROM "group"
		WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]joinGroup, 0)
	for rows.Next() {
		var (
			g          joinGroup
			aliasNames pq.StringArray
		)

		if err := rows.Scan(&g.GroupName, &aliasNames, &g.GroupImg); err != nil {
			return nil, err
		}

		g.AliasNames = []string(aliasNames)
		groups = append(groups, g)
	}

	return groups, rows.Err()
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func writeChannelJoinError(w http.ResponseWriter, status int, message, csrf string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"csrf":  csrf,
		"error": message,
	})
}
