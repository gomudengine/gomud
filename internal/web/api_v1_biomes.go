package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// GET /admin/api/v1/biomes
func apiV1GetBiomesV2(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[[]rooms.BiomeInfo]{
		Success: true,
		Data:    rooms.GetAllBiomesSorted(),
	})
}

// POST /admin/api/v1/biomes
func apiV1CreateBiome(w http.ResponseWriter, r *http.Request) {
	var b rooms.BiomeInfo
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if strings.TrimSpace(b.BiomeId) == "" {
		writeAPIError(w, http.StatusBadRequest, "biomeid is required")
		return
	}
	if err := rooms.CreateBiome(&b); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, APIResponse[rooms.BiomeInfo]{
		Success: true,
		Data:    b,
	})
}

// PATCH /admin/api/v1/biomes/{biomeId}
func apiV1PatchBiome(w http.ResponseWriter, r *http.Request) {
	biomeId := strings.ToLower(strings.TrimSpace(r.PathValue("biomeId")))
	if biomeId == "" {
		writeAPIError(w, http.StatusBadRequest, "biomeId is required")
		return
	}

	existing, ok := rooms.GetBiome(biomeId)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "biome not found: "+biomeId)
		return
	}

	updated := *existing
	// Nil out map fields that must be fully replaced (not merged) by the JSON
	// decode. Go's json.Decoder merges into existing maps, so without this a
	// symbol override removed in the editor (omitted from the payload, or all
	// of them cleared) would keep its old value and reappear on reload.
	updated.SymbolOverrides = nil
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	updated.BiomeId = biomeId

	if err := rooms.SaveBiome(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[rooms.BiomeInfo]{
		Success: true,
		Data:    updated,
	})
}

// DELETE /admin/api/v1/biomes/{biomeId}
func apiV1DeleteBiome(w http.ResponseWriter, r *http.Request) {
	biomeId := strings.ToLower(strings.TrimSpace(r.PathValue("biomeId")))
	if biomeId == "" {
		writeAPIError(w, http.StatusBadRequest, "biomeId is required")
		return
	}
	if err := rooms.DeleteBiome(biomeId); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
