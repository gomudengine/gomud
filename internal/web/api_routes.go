package web

import (
	"net/http"
)

func registerAdminAPIRoutes(mux *http.ServeMux) {
	// Config
	mux.HandleFunc("GET /admin/api/v1/config", RunWithMUDLocked(
		doBasicAuth(apiV1GetConfig),
	))
	mux.HandleFunc("PATCH /admin/api/v1/config", RunWithMUDLocked(
		doBasicAuth(RunInTestMode(apiV1PatchConfig)),
	))

	// StatMods
	mux.HandleFunc("GET /admin/api/v1/statmods", RunWithMUDLocked(
		doBasicAuth(apiV1GetStatMods),
	))

	// Items — static sub-routes must be registered before the wildcard {itemId}
	// pattern so the Go 1.22 ServeMux prefers the more specific match.
	mux.HandleFunc("GET /admin/api/v1/items/types", RunWithMUDLocked(
		doBasicAuth(apiV1GetItemTypes),
	))
	mux.HandleFunc("GET /admin/api/v1/items/attack-messages", RunWithMUDLocked(
		doBasicAuth(apiV1GetItemAttackMessages),
	))
	mux.HandleFunc("GET /admin/api/v1/items", RunWithMUDLocked(
		doBasicAuth(apiV1GetItems),
	))
	mux.HandleFunc("POST /admin/api/v1/items", RunWithMUDLocked(
		doBasicAuth(apiV1CreateItem),
	))
	mux.HandleFunc("GET /admin/api/v1/items/{itemId}/script", RunWithMUDLocked(
		doBasicAuth(apiV1GetItemScript),
	))
	mux.HandleFunc("PUT /admin/api/v1/items/{itemId}/script", RunWithMUDLocked(
		doBasicAuth(apiV1PutItemScript),
	))
	mux.HandleFunc("GET /admin/api/v1/items/{itemId}", RunWithMUDLocked(
		doBasicAuth(apiV1GetItem),
	))
	mux.HandleFunc("PATCH /admin/api/v1/items/{itemId}", RunWithMUDLocked(
		doBasicAuth(apiV1PatchItem),
	))
	mux.HandleFunc("DELETE /admin/api/v1/items/{itemId}", RunWithMUDLocked(
		doBasicAuth(apiV1DeleteItem),
	))

	// Buffs
	mux.HandleFunc("GET /admin/api/v1/buffs", RunWithMUDLocked(
		doBasicAuth(apiV1GetBuffs),
	))
	mux.HandleFunc("GET /admin/api/v1/buffs/{buffId}/script", RunWithMUDLocked(
		doBasicAuth(apiV1GetBuffScript),
	))
	mux.HandleFunc("PUT /admin/api/v1/buffs/{buffId}/script", RunWithMUDLocked(
		doBasicAuth(apiV1PutBuffScript),
	))
	mux.HandleFunc("GET /admin/api/v1/buffs/{buffId}", RunWithMUDLocked(
		doBasicAuth(apiV1GetBuff),
	))
	mux.HandleFunc("PATCH /admin/api/v1/buffs/{buffId}", RunWithMUDLocked(
		doBasicAuth(apiV1PatchBuff),
	))
	mux.HandleFunc("DELETE /admin/api/v1/buffs/{buffId}", RunWithMUDLocked(
		doBasicAuth(apiV1DeleteBuff),
	))

	// Quests
	mux.HandleFunc("GET /admin/api/v1/quests", RunWithMUDLocked(
		doBasicAuth(apiV1GetQuests),
	))
	mux.HandleFunc("PATCH /admin/api/v1/quests", RunWithMUDLocked(
		doBasicAuth(apiV1PatchQuest),
	))
	mux.HandleFunc("DELETE /admin/api/v1/quests/{questId}", RunWithMUDLocked(
		doBasicAuth(apiV1DeleteQuest),
	))
}
