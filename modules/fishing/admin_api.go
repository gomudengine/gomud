package fishing

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (m *FishingModule) apiGetConfig(r *http.Request) (int, bool, any) {
	return http.StatusOK, true, map[string]any{
		"fishingroditems": m.cfg.FishingRodItemIds,
	}
}

func (m *FishingModule) apiPatchConfig(r *http.Request) (int, bool, any) {
	var body struct {
		FishingRodItemIds []int `json:"fishingroditems"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "malformed request body: " + err.Error()}
	}

	if body.FishingRodItemIds == nil {
		body.FishingRodItemIds = []int{}
	}
	m.cfg.FishingRodItemIds = body.FishingRodItemIds
	if err := m.plug.WriteStruct(configKey, m.cfg); err != nil {
		return http.StatusInternalServerError, false, map[string]string{"error": "failed to save config: " + err.Error()}
	}

	return http.StatusOK, true, map[string]string{"status": "saved"}
}

func (m *FishingModule) apiGetCatchables(r *http.Request) (int, bool, any) {
	catchables := m.cfg.Catchables
	if catchables == nil {
		catchables = []CatchableItem{}
	}
	return http.StatusOK, true, catchables
}

func (m *FishingModule) apiAddCatchable(r *http.Request) (int, bool, any) {
	var body CatchableItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "malformed request body: " + err.Error()}
	}

	if body.ItemId < 1 {
		return http.StatusBadRequest, false, map[string]string{"error": "itemid is required"}
	}
	if body.ChanceToCatch < 1 || body.ChanceToCatch > 100 {
		return http.StatusBadRequest, false, map[string]string{"error": "chancetocatch must be between 1 and 100"}
	}

	body.ID = m.nextCatchableId()
	m.cfg.Catchables = append(m.cfg.Catchables, body)

	if err := m.plug.WriteStruct(configKey, m.cfg); err != nil {
		return http.StatusInternalServerError, false, map[string]string{"error": "failed to save: " + err.Error()}
	}

	return http.StatusOK, true, body
}

func (m *FishingModule) apiUpdateCatchable(r *http.Request) (int, bool, any) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return http.StatusBadRequest, false, map[string]string{"error": "valid id is required"}
	}

	var body CatchableItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "malformed request body: " + err.Error()}
	}

	if body.ChanceToCatch < 1 || body.ChanceToCatch > 100 {
		return http.StatusBadRequest, false, map[string]string{"error": "chancetocatch must be between 1 and 100"}
	}

	for i, c := range m.cfg.Catchables {
		if c.ID == id {
			body.ID = id
			m.cfg.Catchables[i] = body
			if err := m.plug.WriteStruct(configKey, m.cfg); err != nil {
				return http.StatusInternalServerError, false, map[string]string{"error": "failed to save: " + err.Error()}
			}
			return http.StatusOK, true, body
		}
	}

	return http.StatusNotFound, false, map[string]string{"error": "catchable not found"}
}

func (m *FishingModule) apiDeleteCatchable(r *http.Request) (int, bool, any) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return http.StatusBadRequest, false, map[string]string{"error": "valid id is required"}
	}

	for i, c := range m.cfg.Catchables {
		if c.ID == id {
			m.cfg.Catchables = append(m.cfg.Catchables[:i], m.cfg.Catchables[i+1:]...)
			if err := m.plug.WriteStruct(configKey, m.cfg); err != nil {
				return http.StatusInternalServerError, false, map[string]string{"error": "failed to save: " + err.Error()}
			}
			return http.StatusOK, true, map[string]string{"status": "deleted"}
		}
	}

	return http.StatusNotFound, false, map[string]string{"error": "catchable not found"}
}
