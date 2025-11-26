# PartyObject

PartyObjects represent collections of actors (users and NPCs) that are grouped together. They provide convenient methods to perform operations on multiple actors at once.

- [PartyObject](#partyobject)
  - [PartyObject.GetMembers() \[\]ActorObject](#partyobjectgetmembers-actorobject)
  - [PartyObject.SendText(msg string)](#partyobjectsendtextmsg-string)
  - [PartyObject.SetResetRoomId(roomId int)](#partyobjectsetresetroomidroomid-int)
  - [PartyObject.GiveQuest(questId string)](#partyobjectgivequestquestid-string)
  - [PartyObject.AddGold(amt int \[, bankAmt int\])](#partyobjectaddgoldamt-int--bankamt-int)
  - [PartyObject.AddHealth(amt int)](#partyobjectaddhealthamt-int)
  - [PartyObject.AddMana(amt int)](#partyobjectaddmanaamt-int)
  - [PartyObject.Command(cmd string \[, waitSeconds float\])](#partyobjectcommandcmd-string--waitseconds-float)
  - [PartyObject.TrainSkill(skillName string, skillLevel int)](#partyobjecttrainskillskillname-string-skilllevel-int)
  - [PartyObject.MoveRoom(destRoomId int)](#partyobjectmoveroomdestroomid-int)
  - [PartyObject.AddEventLog(category string, message string)](#partyobjectaddeventlogcategory-string-message-string)
  - [PartyObject.GiveBuff(buffId int, source string)](#partyobjectgivebuffbuffid-int-source-string)
  - [PartyObject.CancelBuffWithFlag(buffFlag string)](#partyobjectcancelbuffwithflagbuffflag-string)
  - [PartyObject.RemoveBuff(buffId int)](#partyobjectremovebuffbuffid-int)
  - [PartyObject.ChangeAlignment(alignmentChange int)](#partyobjectchangealignmentalignmentchange-int)
  - [PartyObject.LearnSpell(spellId string)](#partyobjectlearnspellspellid-string)
  - [PartyObject.SetHealth(amt int)](#partyobjectsethealthamt-int)
  - [PartyObject.SetAdjective(adj string, addIt bool)](#partyobjectsetadjectiveadj-string-addit-bool)
  - [PartyObject.GiveTrainingPoints(ct int)](#partyobjectgivetrainingpointsct-int)
  - [PartyObject.GiveStatPoints(ct int)](#partyobjectgivestatpointsct-int)
  - [PartyObject.GiveExtraLife()](#partyobjectgiveextralife)
  - [PartyObject.GrantXP(xpAmt int, reason string)](#partyobjectgrantxpxpamt-int-reason-string)
  - [PartyObject.TimerSet(name string, period string)](#partyobjecttimersetname-string-period-string)




## [PartyObject.GetMembers() []ActorObject](/internal/scripting/party_func.go)
Returns an array of all ActorObjects in the party based on the party configuration.

_Note: This includes both party members and their charmed creatures, filtered by presence/absence based on how the party object was created._

## [PartyObject.SendText(msg string)](/internal/scripting/party_func.go)
Sends a message to all members of the party.

|  Argument | Explanation |
| --- | --- |
| msg | the message to send to all party members |

## [PartyObject.SetResetRoomId(roomId int)](/internal/scripting/party_func.go)
Sets the "Reset Room Id" for all user party members (where they will be sent if they log out).

_Note: Only affects user party members, not NPCs. This is only useful if players are being sent to ephemeral chunks._

|  Argument | Explanation |
| --- | --- |
| roomId | The RoomId all party members should be sent to |

## [PartyObject.GiveQuest(questId string)](/internal/scripting/party_func.go)
Grants a quest or progress on a quest to all party members.

|  Argument | Explanation |
| --- | --- |
| questId | The quest identifier string to give, such as `3-start` |

## [PartyObject.AddGold(amt int [, bankAmt int])](/internal/scripting/party_func.go)
Updates how much gold all party members have.

|  Argument | Explanation |
| --- | --- |
| amt | A positive or negative amount of gold to alter each party member's gold by |
| bankAmt (optional) | A positive or negative amount of gold to alter each party member's bank gold by |

## [PartyObject.AddHealth(amt int)](/internal/scripting/party_func.go)
Updates how much health all party members have.

|  Argument | Explanation |
| --- | --- |
| amt | A positive or negative amount of health to alter each party member's health by |

## [PartyObject.AddMana(amt int)](/internal/scripting/party_func.go)
Updates how much mana all party members have.

|  Argument | Explanation |
| --- | --- |
| amt | A positive or negative amount of mana to alter each party member's mana by |

## [PartyObject.Command(cmd string [, waitSeconds float])](/internal/scripting/party_func.go)
Forces all party members to execute a command as if they entered it.

_Note: Don't underestimate the power of this function! Complex and interesting behaviors or interactions can emerge from using it._

|  Argument | Explanation |
| --- | --- |
| cmd | The command to execute such as `look west` or `say goodbye` |
| waitSeconds (optional) | The number of seconds to wait before executing the command |

## [PartyObject.TrainSkill(skillName string, skillLevel int)](/internal/scripting/party_func.go)
Sets a skill level for all party members, if it's greater than what they already have.

|  Argument | Explanation |
| --- | --- |
| skillName | The name of the skill to train, such as `map` or `backstab` |
| skillLevel | The skill level to train to |

## [PartyObject.MoveRoom(destRoomId int)](/internal/scripting/party_func.go)
Quietly moves all party members to a new room.

|  Argument | Explanation |
| --- | --- |
| destRoomId | The room id to move them all to |

## [PartyObject.AddEventLog(category string, message string)](/internal/scripting/party_func.go)
Adds a line to all party members' Event Log (`history`).

|  Argument | Explanation |
| --- | --- |
| category | A short single word category |
| message | A single line describing the event |

## [PartyObject.GiveBuff(buffId int, source string)](/internal/scripting/party_func.go)
Grants all party members a Buff.

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to give them |
| source | The source of the buff, "item", "spell", "trap", "curse", etc. or empty |

## [PartyObject.CancelBuffWithFlag(buffFlag string)](/internal/scripting/party_func.go)
Cancels any buffs that have the flag provided for all party members.

|  Argument | Explanation |
| --- | --- |
| buffFlag | The buff flag to check [see buffspec.go](../buffs/buffspec.go) |

## [PartyObject.RemoveBuff(buffId int)](/internal/scripting/party_func.go)
Remove a buff from all party members without triggering onEnd.

|  Argument | Explanation |
| --- | --- |
| buffId | The ID of the buff to remove |

## [PartyObject.ChangeAlignment(alignmentChange int)](/internal/scripting/party_func.go)
Updates the alignment of all party members by a relative amount. Caps result at -100 to 100.

|  Argument | Explanation |
| --- | --- |
| alignmentChange | The alignment adjustment, from -200 to 200 |

## [PartyObject.LearnSpell(spellId string)](/internal/scripting/party_func.go)
Adds the spell to all party members' spellbooks.

|  Argument | Explanation |
| --- | --- |
| spellId | The ID of the spell |

## [PartyObject.SetHealth(amt int)](/internal/scripting/party_func.go)
Sets all party members' health to a specific amount. If this exceeds their maximum health, sets to their maximum health.

|  Argument | Explanation |
| --- | --- |
| amt | number of hitpoints to set them to |

## [PartyObject.SetAdjective(adj string, addIt bool)](/internal/scripting/party_func.go)
Adds or removes a specific text adjective to all party members' names.

|  Argument | Explanation |
| --- | --- |
| adj | Adjective such as "sleeping", "crying" or "busy" |
| addIt | `true` to add it. `false` to remove it |

## [PartyObject.GiveTrainingPoints(ct int)](/internal/scripting/party_func.go)
Increases training points for all party members.

|  Argument | Explanation |
| --- | --- |
| ct | How many training points to give |

## [PartyObject.GiveStatPoints(ct int)](/internal/scripting/party_func.go)
Increases stat points for all party members.

|  Argument | Explanation |
| --- | --- |
| ct | How many stat points to give |

## [PartyObject.GiveExtraLife()](/internal/scripting/party_func.go)
Increases extra lives by 1 for all party members.

## [PartyObject.GrantXP(xpAmt int, reason string)](/internal/scripting/party_func.go)
Gives experience points to all party members.

|  Argument | Explanation |
| --- | --- |
| xpAmt | How much experience to grant |
| reason | Short reasons such as "combat", "trash cleanup" |

## [PartyObject.TimerSet(name string, period string)](/internal/scripting/party_func.go)
Starts a new Round timer for all party members.

|  Argument | Explanation |
| --- | --- |
| name | A string identifier. Reusing names will overwrite previously assigned names |
| period | How long until the timer expires. `1 real hour`, `1 hour`, etc |