package mudmail

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/GoMudEngine/GoMud/internal/users"
)

// adminSummaryEntry is the JSON shape returned by the list (summary) endpoint.
// It intentionally omits the message body.
type adminSummaryEntry struct {
	UserId     int    `json:"user_id"`
	Username   string `json:"username"`
	FromName   string `json:"from_name"`
	Read       bool   `json:"read"`
	HasGold    bool   `json:"has_gold,omitempty"`
	HasItem    bool   `json:"has_item,omitempty"`
	DateSent   string `json:"date_sent"`    // RFC3339Nano
	DateSentUs int64  `json:"date_sent_us"` // microseconds since Unix epoch
}

// adminMessageEntry is the full JSON shape returned by the body endpoint.
type adminMessageEntry struct {
	UserId     int    `json:"user_id"`
	Username   string `json:"username"`
	FromName   string `json:"from_name"`
	Body       string `json:"body"`
	Gold       int    `json:"gold,omitempty"`
	Read       bool   `json:"read"`
	DateSent   string `json:"date_sent"`    // RFC3339Nano
	DateSentUs int64  `json:"date_sent_us"` // microseconds since Unix epoch
}

// adminSendRequest is the JSON body for the send endpoint.
type adminSendRequest struct {
	FromName string `json:"from_name"`
	Body     string `json:"body"`
	Gold     int    `json:"gold,omitempty"`
	// If UserId is 0 (or omitted), the message is sent to everyone.
	UserId int `json:"user_id,omitempty"`
}

// apiAdminListMudmail handles GET /admin/api/v1/mudmail?user_id=<id>
// Returns summary entries (no body) for the given user.
// user_id is required.
func (m *MudmailModule) apiAdminListMudmail(r *http.Request) (int, bool, any) {
	type result struct {
		Messages []adminSummaryEntry `json:"messages"`
	}

	userIdStr := r.URL.Query().Get("user_id")
	if userIdStr == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "user_id query param is required"}
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil || userId <= 0 {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid user_id"}
	}

	username, inbox := m.inboxForUser(userId)

	entries := make([]adminSummaryEntry, 0, len(inbox))
	for _, msg := range inbox {
		entries = append(entries, toSummaryEntry(userId, username, msg))
	}

	return http.StatusOK, true, result{Messages: entries}
}

// apiAdminGetMudmailBody handles GET /admin/api/v1/mudmail-body/{user_id}/{timestamp}
// Returns the full message data for a single message identified by user_id and
// unix timestamp.
func (m *MudmailModule) apiAdminGetMudmailBody(r *http.Request) (int, bool, any) {
	userIdStr := r.PathValue("user_id")
	timestampStr := r.PathValue("timestamp")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil || userId <= 0 {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid user_id"}
	}
	dateSentUs, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid timestamp"}
	}

	username, inbox := m.inboxForUser(userId)

	for _, msg := range inbox {
		if msg.DateSent.UnixMicro() == dateSentUs {
			return http.StatusOK, true, toAdminEntry(userId, username, msg)
		}
	}

	return http.StatusNotFound, false, map[string]string{"error": "message not found"}
}

// apiAdminSendMudmail handles POST /admin/api/v1/mudmail
// Body: { "from_name": "...", "body": "...", "gold": 0, "user_id": 0 }
// If user_id == 0, sends to everyone.
func (m *MudmailModule) apiAdminSendMudmail(r *http.Request) (int, bool, any) {
	var req adminSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid request body"}
	}
	if req.FromName == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "from_name is required"}
	}
	if req.Body == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "body is required"}
	}

	if req.UserId != 0 {
		m.SendMudMail(req.UserId, req.FromName, req.Body, req.Gold, nil)
		return http.StatusOK, true, map[string]any{"sent_to": req.UserId}
	}

	// Broadcast to everyone.
	msg := Message{
		FromName: req.FromName,
		Body:     req.Body,
		Gold:     req.Gold,
		DateSent: time.Now(),
	}

	onlineIds := map[int]struct{}{}
	for _, u := range users.GetAllActiveUsers() {
		onlineIds[u.UserId] = struct{}{}
		inbox := m.inboxes[u.UserId]
		inbox = append(Inbox{msg}, inbox...)
		m.inboxes[u.UserId] = inbox
		m.save(u.UserId, inbox)
		u.Command(`inbox check`)
	}

	sentCount := len(onlineIds)

	users.SearchOfflineUsers(func(u *users.UserRecord) bool {
		if _, online := onlineIds[u.UserId]; online {
			return true
		}
		inbox := m.load(u.UserId)
		inbox = append(Inbox{msg}, inbox...)
		m.save(u.UserId, inbox)
		sentCount++
		return true
	})

	return http.StatusOK, true, map[string]any{"sent_to_all": true, "recipient_count": sentCount}
}

// apiAdminDeleteMudmail handles DELETE /admin/api/v1/mudmail
// Query params: user_id (int), date_sent_us (int64 microseconds since epoch).
// Deletes the first message for that user whose DateSent microsecond timestamp matches.
func (m *MudmailModule) apiAdminDeleteMudmail(r *http.Request) (int, bool, any) {
	q := r.URL.Query()

	userIdStr := q.Get("user_id")
	dateSentStr := q.Get("date_sent_us")

	if userIdStr == "" || dateSentStr == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "user_id and date_sent_us query params are required"}
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil || userId <= 0 {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid user_id"}
	}

	dateSentUs, err := strconv.ParseInt(dateSentStr, 10, 64)
	if err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid date_sent_us"}
	}

	isOnline := users.GetByUserId(userId) != nil
	var inbox Inbox
	if isOnline {
		inbox = m.inboxes[userId]
	} else {
		inbox = m.load(userId)
	}

	newInbox := make(Inbox, 0, len(inbox))
	deleted := false
	for _, msg := range inbox {
		if !deleted && msg.DateSent.UnixMicro() == dateSentUs {
			deleted = true
			continue
		}
		newInbox = append(newInbox, msg)
	}

	if !deleted {
		return http.StatusNotFound, false, map[string]string{"error": "message not found"}
	}

	if isOnline {
		m.inboxes[userId] = newInbox
	}
	m.save(userId, newInbox)

	return http.StatusOK, true, map[string]any{"deleted": true}
}

// inboxForUser returns the username and inbox for a given userId, preferring
// the in-memory copy for online users and falling back to disk for offline ones.
func (m *MudmailModule) inboxForUser(userId int) (username string, inbox Inbox) {
	if u := users.GetByUserId(userId); u != nil {
		return u.Username, m.inboxes[userId]
	}
	// Offline: find username from index then load from disk.
	idx := users.GetUserIndex()
	if name, ok := idx.FindByUserId(userId); ok {
		username = name
	}
	return username, m.load(userId)
}

// toSummaryEntry converts a Message into the summary shape (no body).
func toSummaryEntry(userId int, username string, msg Message) adminSummaryEntry {
	return adminSummaryEntry{
		UserId:     userId,
		Username:   username,
		FromName:   msg.FromName,
		Read:       msg.Read,
		HasGold:    msg.Gold > 0,
		HasItem:    msg.Item != nil,
		DateSent:   msg.DateSent.UTC().Format(time.RFC3339Nano),
		DateSentUs: msg.DateSent.UnixMicro(),
	}
}

// toAdminEntry converts a Message into the full JSON-friendly adminMessageEntry.
func toAdminEntry(userId int, username string, msg Message) adminMessageEntry {
	return adminMessageEntry{
		UserId:     userId,
		Username:   username,
		FromName:   msg.FromName,
		Body:       msg.Body,
		Gold:       msg.Gold,
		Read:       msg.Read,
		DateSent:   msg.DateSent.UTC().Format(time.RFC3339Nano),
		DateSentUs: msg.DateSent.UnixMicro(),
	}
}
