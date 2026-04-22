package buffs

// Something temporarily attached to a character
// That modifies some aspect of their status
/*
Examples:
Fast Healing - increased natural health recovery for 10 rounds
Poison - add -10 health every round for 5 rounds
*/

type Flag string

const (
	//
	// All Flags must be lowercase
	//
	All Flag = ``

	// Behavioral flags
	NoCombat       Flag = `no-combat`
	NoMovement     Flag = `no-go`
	NoFlee         Flag = `no-flee`
	CancelIfCombat Flag = `cancel-on-combat`
	CancelOnAction Flag = `cancel-on-action`
	CancelOnWater  Flag = `cancel-on-water`

	// Death preventing
	ReviveOnDeath Flag = `revive-on-death`

	// Gear related
	PermaGear   Flag = `perma-gear`
	RemoveCurse Flag = `remove-curse`

	// Harmful flags
	Poison   Flag = `poison`
	Drunk    Flag = `drunk`
	Tripping Flag = `tripping`

	// Useful flags
	Hidden       Flag = `hidden`
	Accuracy     Flag = `accuracy`
	Blink        Flag = `blink`
	EmitsLight   Flag = `lightsource`
	SuperHearing Flag = `superhearing`
	NightVision  Flag = `nightvision`
	Warmed       Flag = `warmed`
	Hydrated     Flag = `hydrated`
	Thirsty      Flag = `thirsty`

	// Flags that reveal things
	SeeHidden Flag = `see-hidden`
	SeeNouns  Flag = `see-nouns`
)

func GetAllFlags() map[Flag]string {
	return map[Flag]string{
		NoCombat:       "Prevents the character from initiating or participating in combat.",
		NoMovement:     "Prevents the character from moving between rooms.",
		NoFlee:         "Prevents the character from fleeing an active combat encounter.",
		CancelIfCombat: "Automatically removes the buff when the character enters combat.",
		CancelOnAction: "Automatically removes the buff when the character takes any action.",
		CancelOnWater:  "Automatically removes the buff when the character enters a water room.",
		ReviveOnDeath:  "Prevents the character from dying once, consuming the buff instead.",
		PermaGear:      "Prevents equipped items from being removed while the buff is active.",
		RemoveCurse:    "Allows cursed items to be removed as if they were not cursed.",
		Poison:         "Marks the character as poisoned, typically dealing periodic health damage.",
		Drunk:          "Marks the character as intoxicated, causing impaired behavior.",
		Tripping:       "Marks the character as tripping, causing hallucinogenic or disorienting effects.",
		Hidden:         "Marks the character as hidden or stealthed, concealing them from others.",
		Accuracy:       "Increases the character's chance to hit and land critical strikes in combat.",
		Blink:          "Grants the character a chance to dodge incoming attacks.",
		EmitsLight:     "Causes the character to emit light, illuminating dark rooms.",
		SuperHearing:   "Grants the character the ability to detect sounds and events in adjacent rooms.",
		NightVision:    "Allows the character to see normally in dark or unlit rooms.",
		Warmed:         "Protects the character from cold environmental effects.",
		Hydrated:       "Marks the character as sufficiently hydrated.",
		Thirsty:        "Marks the character as dehydrated, potentially causing penalties.",
		SeeHidden:      "Allows the character to detect hidden or stealthed entities.",
		SeeNouns:       "Allows the character to see identifying details in rooms.",
	}
}
