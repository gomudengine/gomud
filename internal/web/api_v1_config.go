package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/configs"
)

func apiV1GetConfig(w http.ResponseWriter, r *http.Request) {
	data := configs.GetConfig().AllConfigData()
	writeJSON(w, http.StatusOK, APIResponse[map[string]any]{
		Success: true,
		Data:    data,
	})
}

type patchConfigResult struct {
	Applied  []string `json:"applied"`
	Rejected []string `json:"rejected"`
}

func apiV1PatchConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	result := patchConfigResult{
		Applied:  []string{},
		Rejected: []string{},
	}

	for key, value := range updates {
		err := configs.SetVal(key, value)
		if err == nil {
			result.Applied = append(result.Applied, key)
			continue
		}

		if errors.Is(err, configs.ErrLockedConfig) {
			result.Rejected = append(result.Rejected, key)
			continue
		}

		if errors.Is(err, configs.ErrInvalidConfigName) {
			writeAPIError(w, http.StatusBadRequest, "unknown config key: "+key)
			return
		}

		writeAPIError(w, http.StatusInternalServerError, "internal error applying config")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[patchConfigResult]{
		Success: true,
		Data:    result,
	})
}
