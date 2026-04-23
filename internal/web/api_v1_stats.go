package web

import (
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/util"
)

// memoryEntry is the JSON representation of a single MemoryResult, with the
// unit expressed as a string so callers do not need to know the numeric enum.
type memoryEntry struct {
	Memory uint64 `json:"memory"`
	Count  int    `json:"count"`
	Unit   string `json:"unit"`
}

// GET /admin/api/v1/stats/memory
//
// Returns the output of util.GetMemoryReport() as a JSON object whose keys
// are the section names (e.g. "Go", "Rooms") and whose values are maps of
// metric name to MemoryResult.
func apiV1GetStatsMemory(w http.ResponseWriter, r *http.Request) {
	names, reports := util.GetMemoryReport()

	result := make(map[string]map[string]memoryEntry, len(names))
	for i, name := range names {
		section := make(map[string]memoryEntry, len(reports[i]))
		for metric, res := range reports[i] {
			unit := "bytes"
			if res.Unit == util.UnitCount {
				unit = "count"
			}
			section[metric] = memoryEntry{
				Memory: res.Memory,
				Count:  res.Count,
				Unit:   unit,
			}
		}
		result[name] = section
	}

	writeJSON(w, http.StatusOK, APIResponse[map[string]map[string]memoryEntry]{
		Success: true,
		Data:    result,
	})
}
