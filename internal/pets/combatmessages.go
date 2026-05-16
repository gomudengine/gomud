package pets

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	TokenPetName    = `{petname}`
	TokenDamage     = `{damage}`
	TokenTargetName = `{targetname}`
)

// CombatMessages holds optional custom attack message templates for a pet.
// Each field is a Go text/template-style string supporting the tokens:
//
//	{petname}    - replaced with the pet's display name (may contain ANSI tags)
//	{damage}     - replaced with the integer damage dealt (plain text)
//	{targetname} - replaced with the target's formatted name (may contain ANSI tags)
//
// All four fields are optional. When a field is empty the combat system falls
// back to the built-in default message for that slot.
type CombatMessages struct {
	// ToOwner is sent to the player who owns the pet.
	ToOwner string `yaml:"toowner,omitempty" json:"toowner,omitempty"`
	// ToTarget is sent to the player being attacked (if the target is a player).
	ToTarget string `yaml:"totarget,omitempty" json:"totarget,omitempty"`
	// ToRoom is sent to everyone else in the room observing the attack.
	ToRoom string `yaml:"toroom,omitempty" json:"toroom,omitempty"`
	// Miss is sent to the pet owner when the pet's attack misses.
	Miss string `yaml:"miss,omitempty" json:"miss,omitempty"`
}

// IsEmpty returns true when no custom messages are defined.
func (cm CombatMessages) IsEmpty() bool {
	return cm.ToOwner == `` && cm.ToTarget == `` && cm.ToRoom == `` && cm.Miss == ``
}

// ApplyTokens replaces all recognised tokens in s and returns the result.
// petName and targetName may contain ANSI markup; damage is plain text.
func (cm CombatMessages) ApplyTokens(s, petName string, damage int, targetName string) string {
	s = strings.ReplaceAll(s, TokenPetName, petName)
	s = strings.ReplaceAll(s, TokenDamage, strconv.Itoa(damage))
	s = strings.ReplaceAll(s, TokenTargetName, targetName)
	return s
}

// DefaultCombatMessages returns the built-in fallback messages used when a pet
// type does not define its own. targetType is the ANSI fg colour prefix for the
// target (e.g. "mob" or "user").
func DefaultCombatMessages(targetType string) CombatMessages {
	return CombatMessages{
		ToOwner:  fmt.Sprintf(`{petname} jumps into the fray and deals <ansi fg="damage">{damage} damage</ansi> to <ansi fg="%sname">{targetname}</ansi>!`, targetType),
		ToTarget: `{petname} jumps into the fray and deals <ansi fg="damage">{damage} damage</ansi> to you!`,
		ToRoom:   fmt.Sprintf(`{petname} jumps into the fray and deals <ansi fg="damage">{damage} damage</ansi> to <ansi fg="%sname">{targetname}</ansi>!`, targetType),
		Miss:     fmt.Sprintf(`{petname} lunges at <ansi fg="%sname">{targetname}</ansi> but misses!`, targetType),
	}
}
