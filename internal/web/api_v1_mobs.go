package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

type mobListEntry struct {
	mobs.Mob
	HasScript bool `json:"HasScript"`
}

// GET /admin/api/v1/mobs
func apiV1GetMobs(w http.ResponseWriter, r *http.Request) {
	specs := mobs.GetAllMobSpecs()
	result := make([]mobListEntry, len(specs))
	for i, s := range specs {
		result[i] = mobListEntry{
			Mob:       s,
			HasScript: s.GetScript() != "",
		}
	}
	writeJSON(w, http.StatusOK, APIResponse[[]mobListEntry]{
		Success: true,
		Data:    result,
	})
}

func resolveMobId(w http.ResponseWriter, idOrName string) mobs.MobId {
	if id, err := strconv.Atoi(idOrName); err == nil {
		if spec := mobs.GetMobSpec(mobs.MobId(id)); spec != nil {
			return mobs.MobId(id)
		}
	}
	mobId := mobs.MobIdByName(idOrName)
	if mobId == 0 {
		writeAPIError(w, http.StatusNotFound, "mob not found: "+idOrName)
	}
	return mobId
}

// GET /admin/api/v1/mobs/{mobId}
func apiV1GetMob(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	spec := mobs.GetMobSpec(mobId)
	writeJSON(w, http.StatusOK, APIResponse[*mobs.Mob]{
		Success: true,
		Data:    spec,
	})
}

// POST /admin/api/v1/mobs
func apiV1CreateMob(w http.ResponseWriter, r *http.Request) {
	var spec mobs.Mob
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	newId, err := mobs.CreateNewMobFile(spec, "")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	created := mobs.GetMobSpec(newId)
	writeJSON(w, http.StatusCreated, APIResponse[*mobs.Mob]{
		Success: true,
		Data:    created,
	})
}

// PATCH /admin/api/v1/mobs/{mobId}
func apiV1PatchMob(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	existing := mobs.GetMobSpec(mobId)
	if existing == nil {
		writeAPIError(w, http.StatusNotFound, "mob not found: "+strconv.Itoa(int(mobId)))
		return
	}

	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	updated.MobId = mobId

	if err := mobs.SaveMobSpec(&updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*mobs.Mob]{
		Success: true,
		Data:    &updated,
	})
}

// DELETE /admin/api/v1/mobs/{mobId}
func apiV1DeleteMob(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	if err := mobs.DeleteMobSpec(mobId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// GET /admin/api/v1/mobs/{mobId}/script
func apiV1GetMobScript(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	spec := mobs.GetMobSpec(mobId)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "mob not found")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[map[string]string]{
		Success: true,
		Data:    map[string]string{"script": spec.GetScript()},
	})
}

// PUT /admin/api/v1/mobs/{mobId}/script
func apiV1PutMobScript(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	var body struct {
		Script string `json:"script"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if err := mobs.SaveMobScript(mobId, body.Script); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// PUT /admin/api/v1/mobs/{mobId}/stock
func apiV1PutMobStock(w http.ResponseWriter, r *http.Request) {
	idOrName := r.PathValue("mobId")
	mobId := resolveMobId(w, idOrName)
	if mobId == 0 {
		return
	}

	var entry characters.ShopItem
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if err := mobs.StockMobShop(mobId, entry); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
