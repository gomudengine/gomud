package web

import (
	"net/http"
)

func registerAdminRoutes(mux *http.ServeMux) {
	// Static assets (non-HTML files) under /admin/ require authentication.
	mux.HandleFunc("GET /admin/{file}", doBasicAuth(serveAdminStaticFile))
	mux.HandleFunc("GET /admin/static/{path...}", doBasicAuth(serveAdminStaticFile))

	mux.HandleFunc("GET /admin/", doBasicAuth(RunWithMUDLocked(adminIndex)))
	mux.HandleFunc("GET /admin/https/", doBasicAuth(RunWithMUDLocked(httpsIndex)))

	mux.HandleFunc("GET /admin/config", doBasicAuth(RunWithMUDLocked(adminConfig)))
	mux.HandleFunc("GET /admin/config-api", doBasicAuth(RunWithMUDLocked(adminConfigAPI)))

	mux.HandleFunc("GET /admin/items", doBasicAuth(RunWithMUDLocked(adminItems)))
	mux.HandleFunc("GET /admin/items-rank-weapons", doBasicAuth(RunWithMUDLocked(adminItemsRankWeapons)))
	mux.HandleFunc("GET /admin/items-rank-armor", doBasicAuth(RunWithMUDLocked(adminItemsRankArmor)))
	mux.HandleFunc("GET /admin/items-api", doBasicAuth(RunWithMUDLocked(adminItemsAPI)))
	mux.HandleFunc("GET /admin/items-attack-messages", doBasicAuth(RunWithMUDLocked(adminItemsAttackMessages)))
	mux.HandleFunc("GET /admin/items-attack-messages-api", doBasicAuth(RunWithMUDLocked(adminItemsAttackMessagesAPI)))

	mux.HandleFunc("GET /admin/buffs", doBasicAuth(RunWithMUDLocked(adminBuffs)))
	mux.HandleFunc("GET /admin/buffs-api", doBasicAuth(RunWithMUDLocked(adminBuffsAPI)))

	mux.HandleFunc("GET /admin/quests", doBasicAuth(RunWithMUDLocked(adminQuests)))
	mux.HandleFunc("GET /admin/quests-api", doBasicAuth(RunWithMUDLocked(adminQuestsAPI)))

	mux.HandleFunc("GET /admin/users", doBasicAuth(RunWithMUDLocked(adminUsers)))
	mux.HandleFunc("GET /admin/users-api", doBasicAuth(RunWithMUDLocked(adminUsersAPI)))

	mux.HandleFunc("GET /admin/color-tester", doBasicAuth(RunWithMUDLocked(adminColorTester)))

	mux.HandleFunc("GET /admin/color-aliases", doBasicAuth(RunWithMUDLocked(adminColorAliases)))
	mux.HandleFunc("GET /admin/color-aliases-api", doBasicAuth(RunWithMUDLocked(adminColorAliasesAPI)))

	mux.HandleFunc("GET /admin/colorpatterns", doBasicAuth(RunWithMUDLocked(adminColorPatterns)))
	mux.HandleFunc("GET /admin/colorpatterns-api", doBasicAuth(RunWithMUDLocked(adminColorPatternsAPI)))

	mux.HandleFunc("GET /admin/races", doBasicAuth(RunWithMUDLocked(adminRaces)))
	mux.HandleFunc("GET /admin/races-api", doBasicAuth(RunWithMUDLocked(adminRacesAPI)))

	mux.HandleFunc("GET /admin/keywords", doBasicAuth(RunWithMUDLocked(adminKeywords)))
	mux.HandleFunc("GET /admin/keywords-api", doBasicAuth(RunWithMUDLocked(adminKeywordsAPI)))

	mux.HandleFunc("GET /admin/mobs", doBasicAuth(RunWithMUDLocked(adminMobs)))
	mux.HandleFunc("GET /admin/mobs-api", doBasicAuth(RunWithMUDLocked(adminMobsAPI)))

	mux.HandleFunc("GET /admin/pets", doBasicAuth(RunWithMUDLocked(adminPets)))
	mux.HandleFunc("GET /admin/pets-api", doBasicAuth(RunWithMUDLocked(adminPetsAPI)))
	mux.HandleFunc("GET /admin/pets-ranks", doBasicAuth(RunWithMUDLocked(adminPetsRanks)))

	mux.HandleFunc("GET /admin/mutators", doBasicAuth(RunWithMUDLocked(adminMutators)))
	mux.HandleFunc("GET /admin/mutators-api", doBasicAuth(RunWithMUDLocked(adminMutatorsAPI)))

	mux.HandleFunc("GET /admin/mapper", doBasicAuth(RunWithMUDLocked(adminMapper)))
	mux.HandleFunc("GET /admin/rooms", doBasicAuth(RunWithMUDLocked(adminRooms)))
	mux.HandleFunc("GET /admin/rooms-api", doBasicAuth(RunWithMUDLocked(adminRoomsAPI)))

	mux.HandleFunc("GET /admin/biomes", doBasicAuth(RunWithMUDLocked(adminBiomes)))
	mux.HandleFunc("GET /admin/biomes-api", doBasicAuth(RunWithMUDLocked(adminBiomesAPI)))

	mux.HandleFunc("GET /admin/conversations", doBasicAuth(RunWithMUDLocked(adminConversations)))
	mux.HandleFunc("GET /admin/conversations-api", doBasicAuth(RunWithMUDLocked(adminConversationsAPI)))

	mux.HandleFunc("GET /admin/gametime", doBasicAuth(RunWithMUDLocked(adminGameTime)))
	mux.HandleFunc("GET /admin/gametime-api", doBasicAuth(RunWithMUDLocked(adminGameTimeAPI)))

	mux.HandleFunc("GET /admin/stats", doBasicAuth(RunWithMUDLocked(adminStats)))
	mux.HandleFunc("GET /admin/stats-api", doBasicAuth(RunWithMUDLocked(adminStatsAPI)))

	mux.HandleFunc("GET /admin/telemetry", doBasicAuth(RunWithMUDLocked(adminTelemetry)))
	mux.HandleFunc("GET /admin/telemetry-api", doBasicAuth(RunWithMUDLocked(adminTelemetryAPI)))

	mux.HandleFunc("GET /admin/spells", doBasicAuth(RunWithMUDLocked(adminSpells)))
	mux.HandleFunc("GET /admin/spells-api", doBasicAuth(RunWithMUDLocked(adminSpellsAPI)))

	mux.HandleFunc("GET /admin/audio", doBasicAuth(RunWithMUDLocked(adminAudio)))
	mux.HandleFunc("GET /admin/audio-api", doBasicAuth(RunWithMUDLocked(adminAudioAPI)))

	mux.HandleFunc("GET /admin/scripting-api", doBasicAuth(RunWithMUDLocked(adminScriptingAPI)))

	registerAdminAPIRoutes(mux)
}
