# Pet Scripting

Example scripts:
* [Dog pet script](/_datafiles/world/default/pets/scripts/dog.js)
* [Cat pet script](/_datafiles/world/default/pets/scripts/cat.js)
* [Owl pet script](/_datafiles/world/default/pets/scripts/owl.js)
* [Mule pet script](/_datafiles/world/default/pets/scripts/mule.js)

## Script paths

Pet scripts reside in the same directory as the pet's YAML definition file,
with `.js` replacing `.yaml` in the filename.

For example, the `dog` pet type defined at:

```
_datafiles/world/default/pets/dog.yaml
```

loads its script from:

```
_datafiles/world/default/pets/dog.js
```

## Script scope

Variables defined at the global scope of a pet script are shared across all
owners of that pet type. If you need to store data specific to a single owner,
use the owner's [SetTempData / GetTempData](FUNCTIONS_ACTORS.md#actorobjectsettempdatakey-string-value-any)
or [SetMiscCharacterData / GetMiscCharacterData](FUNCTIONS_ACTORS.md#actorobjectsetmisccharacterdatakey-string-value-any).

## Script functions

The following functions are invoked automatically when defined in a pet script:

---

```
function PetAct(pet PetObject, actor ActorObject, room RoomObject) {
}
```

`PetAct()` is called each round with a probability determined by the pet
type's `RoundActChance` property (0–100). If `RoundActChance` is 0 the
function is never called. If it is 100 it is called every round.

`PetAct` is **not** called while the pet's owner is in combat.

The chance is evaluated by the engine before the function is invoked, so
`PetAct` itself does not need a top-level probability check. Add your own
`RandInt` guard inside the function only if you want behaviour that fires
less often than the configured chance.

There is no return value.

| Argument | Explanation |
| --- | --- |
| pet | [PetObject](FUNCTIONS_PETS.md) — the pet. |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) — the player who owns the pet. |
| room | [RoomObject](FUNCTIONS_ROOMS.md) — the room both are in. |

**Example:**
```javascript
function PetAct(pet, actor, room) {
    // ~5% chance per round to do something visible
    if (RandInt(1, 100) <= 5) {
        room.SendText(pet.NameSimple() + ' sniffs the air curiously.');
    }
}
```

---

```
function onCommand(cmd string, rest string, pet PetObject, actor ActorObject, room RoomObject) {
}
```

`onCommand()` is called whenever the pet's owner types any command.

Returning `true` halts further command processing (the command is consumed).
Returning `false` allows the command to continue through the normal pipeline.

This fires after buff `onCommand` handlers and before item `onCommand`
handlers.

| Argument | Explanation |
| --- | --- |
| cmd | The command word typed by the owner, such as `"look"` or `"attack"`. |
| rest | Everything entered after the command word (may be empty). |
| pet | [PetObject](FUNCTIONS_PETS.md) |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) — the owner. |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

**Example:**
```javascript
function onCommand(cmd, rest, pet, actor, room) {
    if (cmd === 'attack') {
        if (RandInt(1, 4) === 1) {
            room.SendText(pet.NameSimple() + ' growls menacingly.');
        }
    }
    return false;
}
```

---

```
function onCommand_{command}(rest string, pet PetObject, actor ActorObject, room RoomObject) {
}
```

`onCommand_{command}()` is called when the owner types the specific command
named after the underscore.

If a specific handler is defined, the generic `onCommand()` will **not** fire
for that command.

| Argument | Explanation |
| --- | --- |
| rest | Everything entered after the command word (may be empty). |
| pet | [PetObject](FUNCTIONS_PETS.md) |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) — the owner. |
| room | [RoomObject](FUNCTIONS_ROOMS.md) |

**Example:**
```javascript
// Called only when the owner types 'pet' (the pet command)
function onCommand_pet(rest, pet, actor, room) {
    room.SendText(pet.NameSimple() + ' wags its tail happily.');
    return false;
}
```

---

```javascript
function PetLeave(pet, actor, room) {
}
```

`PetLeave()` is called immediately when
[PetObject.GoMissing()](FUNCTIONS_PETS.md#petobjectgomissingrounds-int) is
invoked. Use it to emit a message or trigger effects when the pet disappears.

The pet is already marked as missing when this fires, so `pet.IsMissing()`
returns `true` inside the handler.

There is no return value.

| Argument | Explanation |
| --- | --- |
| pet | [PetObject](FUNCTIONS_PETS.md) — the pet. |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) — the player who owns the pet. |
| room | [RoomObject](FUNCTIONS_ROOMS.md) — the room both are in. |

**Example:**
```javascript
function PetLeave(pet, actor, room) {
    room.SendText(pet.NameSimple() + ' darts into the shadows and disappears!');
}
```

---

```javascript
function PetReturn(pet, actor, room) {
}
```

`PetReturn()` is called the round the pet's `MissingCountdown` reaches zero.
Use it to announce the pet's return or apply any effects.

The countdown has already reached zero when this fires, so `pet.IsMissing()`
returns `false` inside the handler.

There is no return value.

| Argument | Explanation |
| --- | --- |
| pet | [PetObject](FUNCTIONS_PETS.md) — the pet. |
| actor | [ActorObject](FUNCTIONS_ACTORS.md) — the player who owns the pet. |
| room | [RoomObject](FUNCTIONS_ROOMS.md) — the room both are in. |

**Example:**
```javascript
function PetReturn(pet, actor, room) {
    room.SendText(pet.NameSimple() + ' trots back to your side.');
}
```
