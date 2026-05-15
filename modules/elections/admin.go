package elections

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// zoneAuditEntry is the JSON shape returned by GET /admin/api/v1/elections/zones.
type zoneAuditEntry struct {
	Zone          string `json:"zone"`
	OfficialName  string `json:"official_name"`
	OfficialTitle string `json:"official_title"`
	OfficialId    int    `json:"official_id"`
	TaxRate       int    `json:"tax_rate"`
	Coffer        int    `json:"coffer"`
}

// apiGetZones handles GET /admin/api/v1/elections/zones
// Returns all zones with their elected official, tax rate, and coffer balance.
func (m *ElectionsModule) apiGetZones(r *http.Request) (int, bool, any) {
	allZones := rooms.GetAllZoneNames()
	sort.Strings(allZones)

	entries := make([]zoneAuditEntry, 0, len(allZones))
	for _, zone := range allZones {
		zoneKey := strings.ToLower(zone)
		entry := zoneAuditEntry{
			Zone:    zone,
			TaxRate: m.zoneTaxRate(zoneKey),
			Coffer:  m.state.Coffers[zoneKey],
		}
		if w, ok := m.state.Winners[zoneKey]; ok {
			entry.OfficialName = w.CharacterName
			entry.OfficialTitle = w.Title
			entry.OfficialId = w.UserId
		}
		entries = append(entries, entry)
	}

	return http.StatusOK, true, map[string]any{"zones": entries}
}

// apiPatchZone handles PATCH /admin/api/v1/elections/zones/{zone}
// Accepts a JSON body with any combination of: tax_rate (int), coffer (int).
// Validates ranges and writes the changes to module state.
func (m *ElectionsModule) apiPatchZone(r *http.Request) (int, bool, any) {
	zoneName := r.PathValue("zone")
	if zoneName == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "zone is required"}
	}
	zoneKey := strings.ToLower(zoneName)

	var body struct {
		TaxRate *int `json:"tax_rate"`
		Coffer  *int `json:"coffer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, false, map[string]string{"error": "invalid request body"}
	}

	if body.TaxRate == nil && body.Coffer == nil {
		return http.StatusBadRequest, false, map[string]string{"error": "at least one of tax_rate or coffer is required"}
	}

	if body.TaxRate != nil {
		if *body.TaxRate < 0 || *body.TaxRate > 100 {
			return http.StatusBadRequest, false, map[string]string{"error": "tax_rate must be between 0 and 100"}
		}
		m.state.TaxRates[zoneKey] = *body.TaxRate
	}

	if body.Coffer != nil {
		if *body.Coffer < 0 {
			return http.StatusBadRequest, false, map[string]string{"error": "coffer must be non-negative"}
		}
		if *body.Coffer > maxCoffer {
			return http.StatusBadRequest, false, map[string]string{"error": "coffer exceeds maximum of " + strconv.Itoa(maxCoffer)}
		}
		m.state.Coffers[zoneKey] = *body.Coffer
	}

	return http.StatusOK, true, map[string]any{"zone": zoneKey, "updated": true}
}

// apiDeleteZoneOfficial handles DELETE /admin/api/v1/elections/zones/{zone}/official
// Removes the elected official for the named zone.
func (m *ElectionsModule) apiDeleteZoneOfficial(r *http.Request) (int, bool, any) {
	zoneName := r.PathValue("zone")
	if zoneName == "" {
		return http.StatusBadRequest, false, map[string]string{"error": "zone is required"}
	}
	zoneKey := strings.ToLower(zoneName)

	if _, ok := m.state.Winners[zoneKey]; !ok {
		return http.StatusNotFound, false, map[string]string{"error": "no elected official for zone"}
	}

	delete(m.state.Winners, zoneKey)
	return http.StatusOK, true, map[string]any{"zone": zoneKey, "official_removed": true}
}
