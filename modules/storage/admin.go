package storage

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// adminStorageItem is the JSON shape for a single stored item.
type adminStorageItem struct {
	Index   int    `json:"index"`
	ItemId  int    `json:"item_id"`
	Name    string `json:"name"`
	ShortId string `json:"short_id"`
}

// apiAdminGetStorage handles GET /admin/api/v1/storage?user_id=<id>
// Returns the list of stored items for the given user.
func (m *StorageModule) apiAdminGetStorage(r *http.Request) (int, bool, any) {
	type result struct {
		UserId   int                `json:"user_id"`
		Username string             `json:"username"`
		Items    []adminStorageItem `json:"items"`
	}

	userIdStr := r.URL.Query().Get("user_id")
	if userIdStr == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "user_id query param is required"}
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil || userId <= 0 {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid user_id"}
	}

	username, storageItems := m.storageForUser(userId)

	entries := make([]adminStorageItem, 0, len(storageItems))
	for i, itm := range storageItems {
		entries = append(entries, adminStorageItem{
			Index:   i + 1,
			ItemId:  itm.ItemId,
			Name:    itm.Name(),
			ShortId: itm.ShorthandId(),
		})
	}

	return http.StatusOK, true, result{
		UserId:   userId,
		Username: username,
		Items:    entries,
	}
}

// apiAdminDeleteStorageItem handles DELETE /admin/api/v1/storage?user_id=<id>&short_id=<sid>
// Removes the item with the matching short_id from the user's storage.
func (m *StorageModule) apiAdminDeleteStorageItem(r *http.Request) (int, bool, any) {
	// Support both query-param DELETE and JSON body DELETE.
	userIdStr := r.URL.Query().Get("user_id")
	shortId := r.URL.Query().Get("short_id")

	if userIdStr == "" || shortId == "" {
		// Try JSON body.
		var body struct {
			UserId  int    `json:"user_id"`
			ShortId string `json:"short_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if userIdStr == "" {
				userIdStr = strconv.Itoa(body.UserId)
			}
			if shortId == "" {
				shortId = body.ShortId
			}
		}
	}

	if userIdStr == "" || shortId == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "user_id and short_id are required"}
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil || userId <= 0 {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid user_id"}
	}

	isOnline := users.GetByUserId(userId) != nil
	var data StorageData
	if isOnline {
		data = m.storage[userId]
	} else {
		data = m.load(userId)
	}

	var found items.Item
	newItems := make([]items.Item, 0, len(data.Items))
	deleted := false
	for _, itm := range data.Items {
		if !deleted && itm.ShorthandId() == shortId {
			found = itm
			deleted = true
			continue
		}
		newItems = append(newItems, itm)
	}

	if !deleted {
		return http.StatusNotFound, false, map[string]string{"error": "item not found"}
	}

	data.Items = newItems
	if isOnline {
		m.storage[userId] = data
	}
	m.save(userId, data)

	return http.StatusOK, true, map[string]any{
		"deleted":   true,
		"item_id":   found.ItemId,
		"item_name": found.Name(),
	}
}

// storageForUser returns the username and storage items for a given userId,
// preferring the in-memory copy for online users and falling back to disk for offline ones.
func (m *StorageModule) storageForUser(userId int) (username string, storageItems []items.Item) {
	if u := users.GetByUserId(userId); u != nil {
		data := m.storage[userId]
		return u.Username, data.getItems()
	}
	// Offline: find username from index then load from disk.
	idx := users.GetUserIndex()
	if name, ok := idx.FindByUserId(userId); ok {
		username = name
	}
	offlineData := m.load(userId)
	return username, offlineData.getItems()
}
