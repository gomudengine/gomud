# skill.scribe

## Skill Tag & Training

- Skill: `scribe` (`internal/skills/skills.go`)
- Trained at the Dark Acolyte's Chamber (room 160), levels 1-4.

## Overview

`scribe` allows the player to write text in three forms: a portable note item, a public room sign, or a private rune visible only to the scriber. Higher levels unlock the sign and rune forms.

## Preconditions

- Scribe skill level >= 1.
- Command syntax: `scribe <type> <text>`
- Supported types: `note`, `sign`, `rune`.

## Execution Flow

1. Skill level check — rejects if level 0.
2. Argument parsing — splits on whitespace respecting quotes. First token is the scribe type; remainder is the text.
3. Fires a `SkillUsed` event (`skill: scribe`) unconditionally once arguments are parsed.
4. Dispatches to the appropriate handler based on type.

## Scribe Types

### `note` (level 1+)

Creates a new item using item template ID 1 (`items.New(1)`) and sets its blob (text content) to the provided text. The note is stored in the player's backpack.

No cooldown. No length limit enforced in code.

### `sign` (level 2+)

Writes a public sign visible to everyone in the room.

- Rejects if skill level < 2.
- Cooldown: **10 rounds** (keyed `scribe`).
- Text length limit: 50 characters.
- Calls `room.AddSign(text, userId=0, days=7)` — the sign has no owner (userId 0) and expires after 7 game days.
- If a sign already exists in the room, it is replaced and the room is notified.

### `rune` (level 3+)

Writes a private sign visible only to the scriber.

- Rejects if skill level < 3.
- Cooldown: **2 rounds** (keyed `scribe`) — shares the same cooldown key as `sign`.
- Text length limit: 50 characters.
- Calls `room.AddSign(text, userId=user.UserId, days=7)` — the sign is owned by the player (only they can see it) and expires after 7 game days.
- Only the scriber sees the confirmation message; no room broadcast.

## Notes

- Level 4 is listed in the skill description but has no implemented behavior.
- The `sign` and `rune` cooldowns share the same key (`scribe`), so using one blocks the other.
- Notes have no in-game expiry; signs and runes expire after 7 game days.
