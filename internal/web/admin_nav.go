package web

// buildAdminNav returns the full admin navigation, combining hardcoded core
// entries with module-contributed entries.
func buildAdminNav() []WebNavItem {
	nav := []WebNavItem{
		{
			Name:        "Lifeforms",
			Description: "Manage players, NPCs, conversations, pets, races, and character progression.",
			SubMenus: []WebNavItem{
				{
					Name:        "Users",
					Target:      "/admin/users",
					Description: "Manage player accounts, roles, and permissions.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/users", Description: "Browse, search, and edit player account records."},
						{Label: "API Docs", Target: "/admin/users-api", Description: "REST API reference for the users endpoint."},
					},
				},
				{
					Name:        "Mobs",
					Target:      "/admin/mobs",
					Description: "Define and manage non-player character templates.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/mobs", Description: "Browse, create, and edit mob definitions."},
						{Label: "API Docs", Target: "/admin/mobs-api", Description: "REST API reference for the mobs endpoint."},
					},
				},
				{
					Name:        "Conversations",
					Target:      "/admin/conversations",
					Description: "Script dialogue sequences between NPCs.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/conversations", Description: "Browse and edit NPC conversation scripts."},
						{Label: "API Docs", Target: "/admin/conversations-api", Description: "REST API reference for the conversations endpoint."},
					},
				},
				{
					Name:        "Pets",
					Target:      "/admin/pets",
					Description: "Configure tameable companion types and rankings.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/pets", Description: "Browse and edit pet type definitions."},
						{Label: "Rankings", Target: "/admin/pets-ranks", Description: "View pet power rankings across all types."},
						{Label: "API Docs", Target: "/admin/pets-api", Description: "REST API reference for the pets endpoint."},
					},
				},
				{
					Name:        "Races",
					Target:      "/admin/races",
					Description: "Define playable and NPC races with stats and traits.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/races", Description: "Browse and edit race definitions."},
						{Label: "API Docs", Target: "/admin/races-api", Description: "REST API reference for the races endpoint."},
					},
				},
				{
					Name:        "Progression",
					Target:      "/admin/progression",
					Description: "Configure level-up stat gains and experience curves.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/progression", Description: "View and adjust character progression tables."},
					},
				},
			},
		},

		{
			Name:        "Items",
			Description: "Manage item definitions, power tier rankings, and weapon attack messages.",
			SubMenus: []WebNavItem{
				{
					Name:        "Items",
					Target:      "/admin/items",
					Description: "Create and manage all equippable and usable item types.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/items", Description: "Browse, create, and edit item definitions."},
						{Label: "API Docs", Target: "/admin/items-api", Description: "REST API reference for the items endpoint."},
					},
				},
				{
					Name:        "Item Ranks",
					Target:      "/admin/items-rank-weapons",
					Description: "Set power tier rankings for weapons and armor.",
					SubItems: []WebNavSub{
						{Label: "Weapons", Target: "/admin/items-rank-weapons", Description: "View and adjust weapon tier rankings."},
						{Label: "Armor", Target: "/admin/items-rank-armor", Description: "View and adjust armor tier rankings."},
					},
				},
				{
					Name:        "Attack Messages",
					Target:      "/admin/items-attack-messages",
					Description: "Customize flavor text for weapon attack events.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/items-attack-messages", Description: "Browse and edit attack message sets."},
						{Label: "API Docs", Target: "/admin/items-attack-messages-api", Description: "REST API reference for attack messages."},
					},
				},
			},
		},
		{
			Name:        "Rooms",
			Description: "Build and configure the game world: rooms, zones, biomes, and environmental modifiers.",
			SubMenus: []WebNavItem{
				{
					Name:        "Mapper",
					Target:      "/admin/mapper",
					Description: "Visualize and navigate the world map interactively.",
				},
				{
					Name:        "Rooms",
					Target:      "/admin/rooms",
					Description: "Create and configure individual room definitions.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/rooms", Description: "Browse, create, and edit room data."},
						{Label: "API Docs", Target: "/admin/rooms-api", Description: "REST API reference for the rooms endpoint."},
					},
				},
				{
					Name:        "Mutators",
					Target:      "/admin/mutators",
					Description: "Define room-state modifiers that alter environment and spawns.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/mutators", Description: "Browse and edit mutator definitions."},
						{Label: "API Docs", Target: "/admin/mutators-api", Description: "REST API reference for the mutators endpoint."},
					},
				},
				{
					Name:        "Biomes",
					Target:      "/admin/biomes",
					Description: "Configure environmental biome types and their properties.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/biomes", Description: "Browse and edit biome definitions."},
						{Label: "API Docs", Target: "/admin/biomes-api", Description: "REST API reference for the biomes endpoint."},
					},
				},
			},
		},
		{
			Name:        "Buffs",
			Target:      "/admin/buffs",
			Description: "Define status effects applied to characters by spells, items, and rooms.",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/buffs", Description: "Browse, create, and edit buff definitions."},
				{Label: "API Docs", Target: "/admin/buffs-api", Description: "REST API reference for the buffs endpoint."},
			},
		},
		{
			Name:        "Spells",
			Target:      "/admin/spells",
			Description: "Configure castable spells and their scripted behaviors.",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/spells", Description: "Browse, create, and edit spell definitions."},
				{Label: "API Docs", Target: "/admin/spells-api", Description: "REST API reference for the spells endpoint."},
			},
		},
		{
			Name:        "Quests",
			Target:      "/admin/quests",
			Description: "Design multi-step quest chains with token-based progression.",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/quests", Description: "Browse, create, and edit quest definitions."},
				{Label: "API Docs", Target: "/admin/quests-api", Description: "REST API reference for the quests endpoint."},
			},
		},
		{
			Name:        "Colors",
			Description: "Manage color patterns, aliases, and the interactive message crafter.",
			SubMenus: []WebNavItem{
				{
					Name:        "Patterns",
					Target:      "/admin/colorpatterns",
					Description: "Define named multi-color gradient patterns for text.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/colorpatterns", Description: "Browse and edit color pattern definitions."},
						{Label: "API Docs", Target: "/admin/colorpatterns-api", Description: "REST API reference for the color patterns endpoint."},
					},
				},
				{
					Name:        "Aliases",
					Target:      "/admin/color-aliases",
					Description: "Map short color alias names to ANSI color codes.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/color-aliases", Description: "Browse and edit color alias mappings."},
						{Label: "API Docs", Target: "/admin/color-aliases-api", Description: "REST API reference for color aliases."},
					},
				},
				{
					Name:        "Message Crafter",
					Target:      "/admin/color-tester",
					Description: "Interactively compose and preview ANSI-colored messages.",
				},
			},
		},
		{
			Name:        "Keywords",
			Target:      "/admin/keywords",
			Description: "Manage command aliases, help topic keywords, and direction shortcuts.",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/keywords", Description: "Browse and edit keyword and alias definitions."},
				{Label: "API Docs", Target: "/admin/keywords-api", Description: "REST API reference for the keywords endpoint."},
			},
		},
		{
			Name:        "Server",
			Description: "Configure the server, inspect runtime stats, manage time, audio, panels, and scripting.",
			SubMenus: []WebNavItem{
				{
					Name:        "Config",
					Target:      "/admin/config",
					Description: "View and modify live server configuration settings.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/config", Description: "Browse and edit all server configuration values."},
						{Label: "API Docs", Target: "/admin/config-api", Description: "REST API reference for the config endpoint."},
					},
				},
				{
					Name:        "Telemetry",
					Target:      "/admin/telemetry",
					Description: "Query and inspect server-side telemetry event data.",
					SubItems: []WebNavSub{
						{Label: "View / Query", Target: "/admin/telemetry", Description: "Browse and filter recorded telemetry events."},
						{Label: "API Docs", Target: "/admin/telemetry-api", Description: "REST API reference for the telemetry endpoint."},
					},
				},
				{
					Name:        "Stats",
					Target:      "/admin/stats",
					Description: "View real-time memory and runtime statistics.",
					SubItems: []WebNavSub{
						{Label: "View", Target: "/admin/stats", Description: "Display current server memory and runtime metrics."},
						{Label: "API Docs", Target: "/admin/stats-api", Description: "REST API reference for the stats endpoint."},
					},
				},
				{
					Name:        "GameTime",
					Target:      "/admin/gametime",
					Description: "Inspect and control the in-game calendar and clock.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/gametime", Description: "View the current game date and adjust time settings."},
						{Label: "API Docs", Target: "/admin/gametime-api", Description: "REST API reference for the gametime endpoint."},
					},
				},
				{
					Name:        "Audio",
					Target:      "/admin/audio",
					Description: "Manage ambient and event-triggered audio cues.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/audio", Description: "Browse and edit audio cue definitions."},
						{Label: "API Docs", Target: "/admin/audio-api", Description: "REST API reference for the audio endpoint."},
					},
				},
				{
					Name:        "Panels",
					Target:      "/admin/panels",
					Description: "Configure custom UI panels displayed in the web client.",
					SubItems: []WebNavSub{
						{Label: "View / Edit", Target: "/admin/panels", Description: "Browse and edit panel layout definitions."},
						{Label: "API Docs", Target: "/admin/panels-api", Description: "REST API reference for the panels endpoint."},
					},
				},
				{
					Name:        "Scripting",
					Target:      "/admin/scripting",
					Description: "Reference documentation for the server-side scripting system.",
					SubItems: []WebNavSub{
						{Label: "Overview", Target: "/admin/scripting", Description: "Introduction, time periods, and targeting symbols."},
						{Label: "Room Scripts", Target: "/admin/scripting-rooms", Description: "Event hooks for room scripts."},
						{Label: "Mob Scripts", Target: "/admin/scripting-mobs", Description: "Event hooks for mob/NPC scripts."},
						{Label: "Item Scripts", Target: "/admin/scripting-items", Description: "Event hooks for item scripts."},
						{Label: "Buff Scripts", Target: "/admin/scripting-buffs", Description: "Event hooks for buff/status effect scripts."},
						{Label: "Spell Scripts", Target: "/admin/scripting-spells", Description: "Event hooks for spell scripts."},
						{Label: "Pet Scripts", Target: "/admin/scripting-pets", Description: "Event hooks for pet scripts."},
						{Label: "Function Reference", Target: "/admin/scripting-functions", Description: "Full API reference for all scripting objects and functions."},
						{Label: "REST API", Target: "/admin/scripting-api", Description: "REST endpoints for script validation and schema."},
					},
				},
			},
		},
	}
	nav = append(nav, defaultRegistrar.navItems...)
	return nav
}
