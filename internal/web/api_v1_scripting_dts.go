package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/scripting"
)

// handwritten interface declarations for the engine's script object types.
// These correspond to the Go structs ScriptActor, ScriptRoom, ScriptItem,
// ScriptPet, and the schema-driven ContainerObject.
const engineObjectInterfaces = `
declare interface ActorObject {
    UserId(): number;
    InstanceId(): number;
    MobTypeId(): number;
    GetLevel(): number;
    GetStat(statName: string): number;
    GetRace(): string;
    GetTrueRace(): string;
    GetSize(): string;
    IsFormChanged(): boolean;
    ApplyFormChange(raceId: number): boolean;
    RevertFormChange(): boolean;
    SendText(msg: string): void;
    GetName(useShortName?: boolean): string;
    SetTempData(key: string, value: any): void;
    GetTempData(key: string): any;
    SetPermData(key: string, value: any): void;
    GetPermData(key: string): any;
    GetRoomId(): number;
    MoveToRoom(roomId: number): boolean;
    GetHealth(): number;
    GetHealthMax(): number;
    GetMana(): number;
    GetManaMax(): number;
    GetActionPoints(): number;
    GetActionPointsMax(): number;
    GetGold(): number;
    GiveGold(amount: number): void;
    TakeGold(amount: number): number;
    GetAlignment(): number;
    SetAlignment(value: number): void;
    HasBuff(buffId: number): boolean;
    AddBuff(buffId: number): boolean;
    RemoveBuff(buffId: number): boolean;
    GiveItem(item: ItemObject): boolean;
    TakeItem(item: ItemObject): boolean;
    FindItem(itemName: string): ItemObject | null;
    GetItems(): ItemObject[];
    GetWornItems(): ItemObject[];
    IsUser(): boolean;
    IsMob(): boolean;
    IsCharmed(): boolean;
    IsInCombat(): boolean;
    Command(cmd: string, waitTurns?: number): void;
    CommandFlagged(cmd: string, flags: number, waitTurns?: number): void;
    GetSkillLevel(skillName: string): number;
    TrainSkill(skillName: string, amount: number): boolean;
    SetResetRoomId(roomId: number): void;
    GetPet(): PetObject | null;
}

declare interface RoomObject {
    RoomId(): number;
    RoomIdSource(): number;
    SetTempData(key: string, value: any): void;
    GetTempData(key: string): any;
    SetPermData(key: string, value: any): void;
    GetPermData(key: string): any;
    GetItems(): ItemObject[];
    GetStashItems(): ItemObject[];
    DestroyItem(item: ItemObject): void;
    SpawnItem(itemId: number, inStash: boolean): void;
    GetMobs(mobId?: number): ActorObject[];
    GetMob(mobId: number, createIfMissing?: boolean): ActorObject | null;
    GetPlayers(): ActorObject[];
    GetAllActors(): ActorObject[];
    GetContainers(): ContainerObject[];
    GetContainer(name: string): ContainerObject | null;
    SpawnMob(mobId: number): ActorObject | null;
    AddTemporaryExit(exitName: string, exitRoomId: number, duration: string): boolean;
    RemoveTemporaryExit(exitName: string): boolean;
    HasTag(tag: string): boolean;
    SetTag(tag: string): void;
    UnsetTag(tag: string): void;
    GetExits(): string[];
    SendText(msg: string, ...excludeUserIds: number[]): void;
}

declare interface ItemObject {
    ItemId(): number;
    GetUsesLeft(): number;
    SetUsesLeft(amount: number): number;
    AddUsesLeft(amount: number): number;
    GetName(): string;
    GetDescription(): string;
    SetTempData(key: string, value: any): void;
    GetTempData(key: string): any;
    Equals(other: ItemObject): boolean;
}

declare interface PetObject {
    GetName(): string;
    GetType(): string;
    StoredItemCount(): number;
    FindItem(itemName: string): ItemObject | null;
    StoreItem(item: ItemObject): boolean;
    RemoveItem(item: ItemObject): boolean;
    GetItems(): ItemObject[];
    GoMissing(rounds: number): void;
}

declare interface ContainerObject {
    Name(): string;
    HasLock(): boolean;
    IsLocked(): boolean;
    Lock(): void;
    Unlock(): void;
    GetItems(): ItemObject[];
    FindItem(itemName: string): ItemObject | null;
    AddItem(item: ItemObject): boolean;
    RemoveItem(item: ItemObject): boolean;
    GetGold(): number;
    AddGold(amount: number): void;
    RemoveGold(amount: number): number;
    Count(itemId: number): number;
    IsTemporary(): boolean;
}

declare interface CommandEventDetails {
    sourceId: number;
    sourceType: string;
}

declare interface GiveEventDetails {
    sourceId: number;
    sourceType: string;
    item: ItemObject | null;
    gold: number;
}

declare interface ShowEventDetails {
    sourceId: number;
    sourceType: string;
    item: ItemObject | null;
}

declare interface AskEventDetails {
    sourceId: number;
    sourceType: string;
    askText: string;
}

declare interface HurtEventDetails {
    sourceId: number;
    sourceType: string;
    damage: number;
    crit: boolean;
}

declare interface DieEventDetails {
    sourceId: number;
    sourceType: string;
    attackerCount: number;
}

declare interface PathEventDetails {
    sourceId: number;
    sourceType: string;
    status: string;
}

declare interface PanelLayoutObject {
    [key: string]: any;
}

declare interface MatchResult {
    found: boolean;
    exact: string;
    close: string;
}

declare interface GameDate {
    Year: number;
    Month: number;
    Week: number;
    Day: number;
    Hour: number;
    Hour24: number;
    Minute: number;
    AmPm: string;
    Night: boolean;
    DayStart: number;
    NightStart: number;
}
`

// goTypeToTS converts a schema type string to its TypeScript equivalent.
func goTypeToTS(t string) string {
	switch t {
	case "number", "string", "boolean", "void", "any", "object":
		return t
	case "number[]":
		return "number[]"
	case "string[]":
		return "string[]"
	case "boolean[]":
		return "boolean[]"
	case "ActorObject[]":
		return "ActorObject[]"
	case "ItemObject[]":
		return "ItemObject[]"
	case "number|string":
		return "number | string"
	case "varies":
		return "ActorObject | ActorObject[] | string"
	case "RoomObject":
		return "RoomObject"
	case "ActorObject":
		return "ActorObject"
	case "ItemObject":
		return "ItemObject"
	case "PetObject":
		return "PetObject"
	case "ContainerObject":
		return "ContainerObject"
	case "PanelLayoutObject":
		return "PanelLayoutObject"
	default:
		// Pass through unknown types (e.g. union types already written as TS)
		return t
	}
}

// buildParamList renders a TypeScript parameter list from a slice of ScriptFuncParam.
func buildParamList(params []scripting.ScriptFuncParam) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		name := p.Name
		tsType := goTypeToTS(p.Type)
		// Variadic params use "..." prefix in the name field
		if strings.HasPrefix(name, "...") {
			parts = append(parts, fmt.Sprintf("...%s: %s[]", name[3:], goTypeToTS(p.Type)))
			continue
		}
		// Optional params use "?" suffix in the name field
		optional := strings.HasSuffix(name, "?")
		if optional {
			name = name[:len(name)-1]
		}
		if optional {
			parts = append(parts, fmt.Sprintf("%s?: %s", name, tsType))
		} else {
			parts = append(parts, fmt.Sprintf("%s: %s", name, tsType))
		}
	}
	return strings.Join(parts, ", ")
}

// GET /admin/api/v1/scripting/types.d.ts?type=<scriptType>
func apiV1GetScriptingTypesDts(w http.ResponseWriter, r *http.Request) {
	scriptType := r.URL.Query().Get("type")
	schema := scripting.GetScriptFunctionsSchema()

	var sb strings.Builder

	// Engine object interfaces (hand-authored, stable)
	sb.WriteString(engineObjectInterfaces)
	sb.WriteByte('\n')

	// Global engine function declarations (generated from schema)
	sb.WriteString("// Global engine functions\n")
	for _, fn := range schema.EngineFunctions {
		params := buildParamList(fn.Params)
		retType := "void"
		if fn.ReturnType != "" {
			retType = goTypeToTS(fn.ReturnType)
		}
		if fn.Description != "" {
			sb.WriteString(fmt.Sprintf("/** %s */\n", fn.Description))
		}
		sb.WriteString(fmt.Sprintf("declare function %s(%s): %s;\n", fn.Name, params, retType))
	}

	// Script-type-specific event callback declarations
	if scriptType != "" {
		typeDef, ok := schema.ScriptTypes[scriptType]
		if ok {
			sb.WriteString(fmt.Sprintf("\n// Event callbacks for script type: %s\n", scriptType))
			for _, fn := range typeDef.Functions {
				if fn.Dynamic != nil {
					placeholder := fn.Dynamic.Placeholder
					exampleName := strings.ReplaceAll(fn.Name, placeholder, "commandName")
					params := buildParamList(fn.Params)
					retType := "boolean | void"
					if fn.ReturnType != "" {
						retType = goTypeToTS(fn.ReturnType)
					}
					if fn.Description != "" {
						sb.WriteString(fmt.Sprintf("/** %s (dynamic: replace %s with the actual command word) */\n", fn.Description, placeholder))
					}
					sb.WriteString(fmt.Sprintf("declare function %s(%s): %s;\n", exampleName, params, retType))
					continue
				}
				params := buildParamList(fn.Params)
				retType := "void"
				if fn.ReturnType != "" {
					retType = goTypeToTS(fn.ReturnType)
				}
				if fn.Description != "" {
					sb.WriteString(fmt.Sprintf("/** %s */\n", fn.Description))
				}
				sb.WriteString(fmt.Sprintf("declare function %s(%s): %s;\n", fn.Name, params, retType))
			}

			// Emit @callback JSDoc typedefs for every non-dynamic callback.
			// When the user writes a function with a matching name, Monaco's
			// JS checker resolves the @callback type and infers parameter types
			// inside the function body, enabling member completions on typed params.
			sb.WriteString(fmt.Sprintf("\n// JSDoc callback typedefs for script type: %s\n", scriptType))
			for _, fn := range typeDef.Functions {
				if fn.Dynamic != nil {
					continue
				}
				cbName := strings.ToUpper(fn.Name[:1]) + fn.Name[1:] + "Callback"
				sb.WriteString(fmt.Sprintf("/**\n * @callback %s\n", cbName))
				for _, p := range fn.Params {
					paramName := p.Name
					if strings.HasPrefix(paramName, "...") {
						paramName = paramName[3:]
					}
					if strings.HasSuffix(paramName, "?") {
						paramName = paramName[:len(paramName)-1]
					}
					tsType := goTypeToTS(p.Type)
					sb.WriteString(fmt.Sprintf(" * @param {%s} %s - %s\n", tsType, paramName, p.Description))
				}
				retType := "void"
				if fn.ReturnType != "" {
					retType = goTypeToTS(fn.ReturnType)
				}
				sb.WriteString(fmt.Sprintf(" * @returns {%s}\n */\n", retType))
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, sb.String())
}
