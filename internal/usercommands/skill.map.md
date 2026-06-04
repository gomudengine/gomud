# skill.map

## Skill Tag & Training

- Skill: `map` (`internal/skills/skills.go`)
- Levels 1-4 (and an implicit admin level > 4).

## Overview

`map` generates and displays an ASCII map of the current zone or a named zone. Higher skill levels produce larger maps and lower zoom levels (more detail). The map grays out rooms the player has not yet visited. Party members and their charmed mobs are shown as overlaid symbols.

## Preconditions

- Map skill level >= 1.
- Cooldown: **1 round** per use (keyed `map`).

## Map Dimensions by Skill Level

The mapper uses `ZoomLevel = 5 - skillLevel`, meaning higher skill levels produce a closer zoom (more rooms visible):

| Level | ZoomLevel | Approximate visible area |
|---|---|---|
| 1 | 4 | 5×5 |
| 2 | 3 | 9×7 |
| 3 | 2 | 13×9 |
| 4 | 1 | 17×9 |

The canvas is fixed at 65×21 characters for levels 1-4. Level > 4 (admin) scales to the client's reported screen dimensions.

## Sprawl Capacity

```
sprawlCap = skillLevel + (Smarts(adj) / 4)
```

Accessible via `map sprawl` which prints the player's current sprawl capacity. Sprawl affects how many rooms the mapper traverses from the origin.

## Zone Selection

- No argument: maps the player's current zone, centered on their current room.
- Named zone argument: attempts to find the zone by name via `rooms.FindZoneName`. If a different zone is specified, the map is centered on that zone's root room.
- `wide` argument: requires level 4; enables the wide map variant (currently uses the same rendering path as normal).

## Pre-made Maps

Before generating a dynamic map, the command checks for a static template at `maps/<sanitized-zone-name>`. If found, it is rendered directly and the dynamic generation is skipped.

## Dynamic Map Generation

1. Retrieves the zone mapper via `mapper.GetMapper(roomId)`.
2. Builds a `mapper.Config` with zoom level, dimensions, user ID, and visited-room set.
3. For levels 1-4, visited rooms are loaded from `user.Character.ZonesVisited` (the persistent room-visit bitset). Unvisited rooms are rendered differently by the mapper.
4. Party members' current rooms are overlaid with `☺` ("Party Member") symbols. Their charmed mobs are overlaid with `☹` ("Friend").
5. The player's current room is overlaid with `@` ("You").
6. Calls `zMapper.GetLimitedMap(roomId, c)` to produce the render output.
7. Each cell is converted to an ANSI-tagged string using the cell's foreground/background color and symbol.
8. Zone completion percentage is calculated via `user.Character.ZoneVisitPercent(zone, zCfg.RoomIds)` if a zone config exists.
9. The final map is rendered through the `maps/map` template with title, legend, borders, and completion percentage.

## Admin Mode (skill level > 4)

Admin maps scale to the full client screen. Additionally:
- Rooms with hostile/fighting mobs are overlaid with `☠` ("Mob").
- Rooms with non-hostile mobs are overlaid with `☺` ("NPC").
- Rooms with players are overlaid with `☺` ("Player").

## Notes

- The legend is built from `keywords.GetAllLegendAliases(room.Zone)`, providing zone-specific symbol descriptions.
- The map is purely informational and has no effect on game state.
