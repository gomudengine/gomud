# PetObject

PetObject represents a pet owned by a player. It is obtained by calling
[ActorObject.GetPet()](FUNCTIONS_ACTORS.md#actorobjectgetpet) and returns
`null` when the actor has no pet.

The object is a **live reference** into the owner's character data. Mutations
take effect immediately and are persisted the next time the character is saved.

- [PetObject](#petobject)
  - [PetObject.Type() string](#petobjecttype-string)
  - [PetObject.Name() string](#petobjectname-string)
  - [PetObject.NameSimple() string](#petobjectnamesimple-string)
  - [PetObject.SetName(name string)](#petobjectsetnamename-string)
  - [PetObject.Level() int](#petobjectlevel-int)
  - [PetObject.Food() string](#petobjectfood-string)
  - [PetObject.FoodLevel() int](#petobjectfoodlevel-int)
  - [PetObject.Feed()](#petobjectfeed)
  - [PetObject.Starve()](#petobjectstarve)
  - [PetObject.GetStatMod(statName string) int](#petobjectgetstatmodstatname-string-int)
  - [PetObject.GetCapacity() int](#petobjectgetcapacity-int)
  - [PetObject.ItemCount() int](#petobjectitemcount-int)
  - [PetObject.HasScript() bool](#petobjecthasscript-bool)

---

## [PetObject.Type() string](/internal/scripting/pet_func.go)
Returns the pet's type identifier, such as `"dog"`, `"cat"`, `"owl"`, or `"mule"`.

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null && pet.Type() === 'dog') {
    room.SendText(pet.NameSimple() + ' barks!');
}
```

---

## [PetObject.Name() string](/internal/scripting/pet_func.go)
Returns the pet's full display name, including ANSI colour tags and any hunger
indicator such as `(Hungry)` or `(Starving)`.

Use this when sending the name to players as part of room or character output.
Use [NameSimple()](#petobjectnamesimple-string) when you need a plain string
for comparisons or concatenation.

---

## [PetObject.NameSimple() string](/internal/scripting/pet_func.go)
Returns the plain text name of the pet with no colour tags or hunger indicator.
Falls back to the type identifier if the player has not given the pet a custom
name.

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null) {
    room.SendText(pet.NameSimple() + ' trots along behind you.');
}
```

---

## [PetObject.SetName(name string)](/internal/scripting/pet_func.go)
Renames the pet. The change is reflected immediately in `Name()` and
`NameSimple()`.

Pass an empty string to clear the custom name and revert the display name to
the pet's type identifier.

| Argument | Explanation |
| --- | --- |
| name | The new plain text name, or `""` to clear it. |

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null && pet.NameSimple() === pet.Type()) {
    pet.SetName('Biscuit');
    actor.SendText('You name your pet Biscuit.');
}
```

---

## [PetObject.Level() int](/internal/scripting/pet_func.go)
Returns the pet's current level, from 1 (minimum) to 10 (maximum).

Pet level increases when the pet is well-fed at the daily tick and decreases
when the pet is starving.

---

## [PetObject.Food() string](/internal/scripting/pet_func.go)
Returns the pet's current hunger state as a human-readable string.

| Value | Meaning |
| --- | --- |
| `"Starving"` | Food level 0 — pet will lose a level at the next daily tick |
| `"Hungry"` | Food level 1 |
| `"Satisfied"` | Food level 2 |
| `"Full"` | Food level 3 — pet will gain a level at the next daily tick |

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null && pet.Food() === 'Starving') {
    actor.SendText(pet.NameSimple() + ' looks at you with desperate eyes.');
}
```

---

## [PetObject.FoodLevel() int](/internal/scripting/pet_func.go)
Returns the pet's raw hunger value: `0` (Starving) through `3` (Full).

Useful when you need a numeric comparison rather than a string.

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null && pet.FoodLevel() < 2) {
    actor.SendText(pet.NameSimple() + ' nudges you hopefully.');
}
```

---

## [PetObject.Feed()](/internal/scripting/pet_func.go)
Increases the pet's hunger level by one step, up to a maximum of 3 (Full).
Has no effect if the pet is already full.

_Note: This does not consume any item from the player's inventory. Use it when
a script wants to feed the pet as a side-effect of some other action._

---

## [PetObject.Starve()](/internal/scripting/pet_func.go)
Decreases the pet's hunger level by one step, down to a minimum of 0
(Starving). Has no effect if the pet is already starving.

---

## [PetObject.GetStatMod(statName string) int](/internal/scripting/pet_func.go)
Returns the stat modifier the pet currently grants its owner for the named
stat. Returns `0` if the pet provides no modifier for that stat at its current
level.

Stat modifiers are defined per ability level in the pet's YAML definition and
scale up as the pet levels.

| Argument | Explanation |
| --- | --- |
| statName | A stat name such as `"strength"`, `"speed"`, `"smarts"`, `"vitality"`, `"mysticism"`, or `"perception"`. |

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null) {
    var speedBonus = pet.GetStatMod('speed');
    if (speedBonus > 0) {
        actor.SendText(pet.NameSimple() + ' makes you feel quicker. (+' + speedBonus + ' speed)');
    }
}
```

---

## [PetObject.GetCapacity() int](/internal/scripting/pet_func.go)
Returns the number of items the pet can carry at its current level. Returns `0`
if the pet type has no carry ability (e.g. a cat or owl).

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null && pet.GetCapacity() > 0) {
    actor.SendText(pet.NameSimple() + ' can carry ' + pet.GetCapacity() + ' item(s).');
}
```

---

## [PetObject.ItemCount() int](/internal/scripting/pet_func.go)
Returns the number of items the pet is currently carrying.

**Example:**
```javascript
var pet = actor.GetPet();
if (pet !== null) {
    var free = pet.GetCapacity() - pet.ItemCount();
    actor.SendText(pet.NameSimple() + ' has ' + free + ' free slot(s).');
}
```

---

## [PetObject.HasScript() bool](/internal/scripting/pet_func.go)
Returns `true` if this pet type has a script file on disk.

Useful in generic scripts that want to check whether a pet will respond to
events before attempting to trigger them.
