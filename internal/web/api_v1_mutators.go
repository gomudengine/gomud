package web

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/mutators"
)

// GET /admin/api/v1/mutators
func apiV1GetMutators(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[[]mutators.MutatorSpec]{
		Success: true,
		Data:    mutators.GetAllMutatorSpecs(),
	})
}

// POST /admin/api/v1/mutators
func apiV1CreateMutator(w http.ResponseWriter, r *http.Request) {
	var spec mutators.MutatorSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	newId, err := mutators.CreateNewMutatorSpec(spec)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	created := mutators.GetMutatorSpec(newId)
	writeJSON(w, http.StatusCreated, APIResponse[*mutators.MutatorSpec]{
		Success: true,
		Data:    created,
	})
}

// GET /admin/api/v1/mutators/{mutatorId}
func apiV1GetMutator(w http.ResponseWriter, r *http.Request) {
	mutatorId := r.PathValue("mutatorId")

	spec := mutators.GetMutatorSpec(mutatorId)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "mutator not found: "+mutatorId)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*mutators.MutatorSpec]{
		Success: true,
		Data:    spec,
	})
}

// PATCH /admin/api/v1/mutators/{mutatorId}
func apiV1PatchMutator(w http.ResponseWriter, r *http.Request) {
	mutatorId := r.PathValue("mutatorId")

	existing := mutators.GetMutatorSpec(mutatorId)
	if existing == nil {
		writeAPIError(w, http.StatusNotFound, "mutator not found: "+mutatorId)
		return
	}

	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	updated.MutatorId = mutatorId

	if err := mutators.SaveMutatorSpec(&updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*mutators.MutatorSpec]{
		Success: true,
		Data:    &updated,
	})
}

// DELETE /admin/api/v1/mutators/{mutatorId}
func apiV1DeleteMutator(w http.ResponseWriter, r *http.Request) {
	mutatorId := r.PathValue("mutatorId")

	if err := mutators.DeleteMutatorSpec(mutatorId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
