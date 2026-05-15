# Election Tax Plan

## Assessment

The elections module is well-structured for this change. The key mechanisms already in place are:

- `events.Purchase` listener in `onPurchase` - already receives every shop purchase with a room ID and cost, and routes a percentage to the zone coffer. This is where the player-facing tax charge needs to be added.
- `usercommands.OnShopList` hook (`internal/usercommands/shophooks.go`) - a `util.Hook[ShopListRequest]` fired before every shop listing and every buy attempt. This is where the tax annotation on shop listings belongs.
- `rooms.OnRoomLook` hook - already used by the elections module to inject room alerts. The tax rate notice for a zone can be injected here.
- `ElectionsState` - the persisted module state struct; a `TaxRates map[string]int` field can be added here (zone key -> rate 0-100, default 1).

### What needs to change in core code

The `ShopListRequest` struct (`internal/usercommands/shophooks.go`) currently has no room context. The `list` and `buy` commands both have a `*rooms.Room` argument but do not thread it into the hook payload. To let the elections module know which zone a shop is in, `ShopListRequest` needs a `Room *rooms.Room` field populated at both call sites.

This is a small, additive change to a hook struct that already exists for exactly this purpose. It does not break any existing handlers (the new field is simply nil for handlers that do not use it, but in practice it will always be set since `list` and `buy` always have a room). Removing the elections module leaves this field unused but harmless.

### What stays entirely inside the elections module

- Tax rate storage and defaults
- The `election taxrate` sub-command for the elected official
- The `OnShopList` handler that annotates displayed prices with the zone tax
- The `OnShopList` handler that charges the additional tax gold during a buy
- The `OnRoomLook` handler addition that shows the current tax rate in shop rooms
- All persistence via the existing `ElectionsState` save/load path

---

## Implementation Plan

### Step 1 - Add `Room` to `ShopListRequest` (core change)

**File:** `internal/usercommands/shophooks.go`

Add a `Room *rooms.Room` field to `ShopListRequest`. This is the only required core change. The field carries the room the buyer is standing in, which the elections module needs to determine the zone and look up the tax rate.

```go
type ShopListRequest struct {
    Stock      characters.Shop
    Buyer      *users.UserRecord
    SellerMob  *mobs.Mob
    SellerUser *users.UserRecord
    Room       *rooms.Room  // add this
}
```

**Files:** `internal/usercommands/list.go` and `internal/usercommands/buy.go`

Populate `Room` at all four `OnShopList.Fire(...)` call sites (two in `list.go`, two in `buy.go`). Each call site already has a `room *rooms.Room` parameter in scope.

```go
shopReq := OnShopList.Fire(ShopListRequest{
    Stock:     mob.Character.Shop.GetInstock(),
    Buyer:     user,
    SellerMob: mob,
    Room:      room,  // add this
})
```

No other core files require changes.

---

### Step 2 - Add tax rate state to `ElectionsState`

**File:** `modules/elections/elections.go`

Add `TaxRates map[string]int` to `ElectionsState` (zone key -> rate, 0-100).

```go
type ElectionsState struct {
    ActiveElection *Election         `yaml:"activeelection,omitempty"`
    Winners        map[string]Winner `yaml:"winners,omitempty"`
    Coffers        map[string]int    `yaml:"coffers,omitempty"`
    TaxRates       map[string]int    `yaml:"taxrates,omitempty"`
}
```

Initialize the map in `load()` alongside the existing nil-checks:

```go
if m.state.TaxRates == nil {
    m.state.TaxRates = make(map[string]int)
}
```

Add a helper method that returns the effective tax rate for a zone, defaulting to 1:

```go
func (m *ElectionsModule) zoneTaxRate(zoneKey string) int {
    if rate, ok := m.state.TaxRates[zoneKey]; ok {
        return rate
    }
    return 1
}
```

---

### Step 3 - Add the `election taxrate` sub-command

**File:** `modules/elections/elections.go`

Extend `electionAdminCommand` with a `taxrate` case. Only the current elected official of the zone may set the rate. Admins may set any zone's rate by specifying the zone name as an additional argument.

Usage:
- `election taxrate <0-100>` - sets the tax rate for the zone the official is currently in
- `election taxrate <zonename> <0-100>` - admin override for any zone

```
case `taxrate`:
    // parse args, validate 0-100
    // check that caller is admin or is the winner for the zone
    // set m.state.TaxRates[zoneKey] = rate
    // confirm to caller
```

The rate of 0 means no tax is charged. The rate is capped at 100.

---

### Step 4 - Annotate shop listings with the tax rate

#### 4a - Add `OnShopListRendered` hook (core change)

**File:** `internal/usercommands/shophooks.go`

Add a second hook that fires once per seller after all of that seller's tables have been sent to the buyer. It reuses the same `ShopListRequest` type — no new struct needed.

```go
// OnShopListRendered is fired once per seller after all shop tables for that
// seller have been sent to the buyer. Modules register handlers here to append
// additional text (e.g. zone tax notices) after the listing.
var OnShopListRendered util.Hook[ShopListRequest]
```

**File:** `internal/usercommands/list.go`

Fire `OnShopListRendered` once per seller, after all of that seller's category tables have been sent. This is inside the mob loop and the player loop, after the last `user.SendText(...)` for that seller. Pass the same `ShopListRequest` that was already built for `OnShopList` (including the `Room` field added in Step 1).

```go
// after the last SendText for this seller:
OnShopListRendered.Fire(shopReq)
```

This hook is not fired in `buy.go` — it is a display-only extension point.

#### 4b - Register the tax notice handler in the elections module

**File:** `modules/elections/elections.go`

Register a handler on `usercommands.OnShopListRendered` during `init()`. The handler checks the zone tax rate and sends a single line to the buyer after the listing:

```go
usercommands.OnShopListRendered.Register(func(r usercommands.ShopListRequest) usercommands.ShopListRequest {
    if r.Room == nil || r.Buyer == nil {
        return r
    }
    zoneKey := strings.ToLower(r.Room.Zone)
    rate := m.zoneTaxRate(zoneKey)
    r.Buyer.SendText(fmt.Sprintf(
        `<ansi fg="yellow">Zone tax rate: <ansi fg="white-bold">%d%%</ansi> (charged on top of listed prices)</ansi>`,
        rate,
    ))
    return r
})
```

#### 4c - Charge the tax at purchase time via `OnShopList`

Register a separate handler on `usercommands.OnShopList` (which fires in both `list.go` and `buy.go`) to mutate item prices to the tax-inclusive amount. The handler runs on the stock copy — the original shop inventory is never modified.

```go
usercommands.OnShopList.Register(func(r usercommands.ShopListRequest) usercommands.ShopListRequest {
    if r.Room == nil {
        return r
    }
    zoneKey := strings.ToLower(r.Room.Zone)
    rate := m.zoneTaxRate(zoneKey)
    if rate == 0 {
        return r
    }
    for i, item := range r.Stock {
        if item.Price > 0 {
            r.Stock[i].Price = item.Price + (item.Price * rate / 100)
        }
    }
    return r
})
```

Because `buy.go` fires `OnShopList` before resolving the final price, this handler ensures the buyer is charged the tax-inclusive amount automatically. The coffer contribution in `onPurchase` operates on `evt.Cost` (the amount actually charged), so it automatically receives its share of the tax-inclusive price with no further changes.

---

### Step 5 - Show the tax rate in room alerts for shop rooms

**File:** `modules/elections/elections.go`

Extend `onRoomLook` to inject a tax rate notice in any room in a governed zone (one with an elected winner). Since the room alert is zone-wide information and the `RoomTemplateDetails` carries the zone name, this requires no additional tags:

```go
// inside onRoomLook, after existing tag checks:
zoneKey := strings.ToLower(d.Zone)
if _, hasWinner := m.state.Winners[zoneKey]; hasWinner {
    rate := m.zoneTaxRate(zoneKey)
    d.RoomAlerts = append(d.RoomAlerts,
        fmt.Sprintf(`<ansi fg="yellow">Zone tax rate: <ansi fg="white-bold">%d%%</ansi></ansi>`, rate),
    )
}
```

---

### Step 6 - Update help text

**File:** `modules/elections/files/datafiles/templates/help/elections.template`

Add a section describing the tax rate mechanic: what it is, the default (1%), the range (0-100%), and the command to change it (`election taxrate <rate>`).

---

## Summary of file changes

| File | Change type | Reason |
|---|---|---|
| `internal/usercommands/shophooks.go` | Add field + add hook | `Room *rooms.Room` on `ShopListRequest`; `OnShopListRendered` hook |
| `internal/usercommands/list.go` | Populate field + fire hook | Pass `room` at both `OnShopList.Fire` call sites; fire `OnShopListRendered` per seller |
| `internal/usercommands/buy.go` | Populate field | Pass `room` at both `OnShopList.Fire` call sites |
| `modules/elections/elections.go` | Feature addition | Tax rate state, command, `OnShopList` price handler, `OnShopListRendered` notice handler, room alert |
| `modules/elections/files/datafiles/templates/help/elections.template` | Docs update | Describe tax rate mechanic |

Total core changes: 3 files, all additive. Deleting the elections module leaves `Room` and `OnShopListRendered` as unused but harmless additions.

---

## Design decisions and tradeoffs

**Why a separate `OnShopListRendered` hook rather than putting the notice in `OnShopList`?**
`OnShopList` fires before rendering — its purpose is stock mutation. Sending text to the player from inside that handler would inject the notice before the table, not after. `OnShopListRendered` fires after all tables for a seller have been sent, which is exactly when the tax line should appear. The two hooks serve distinct purposes: one mutates data, the other appends output.

**Why default to 1% rather than 0%?**
The requirement specifies 1% as the default. A rate of 0 is a valid explicit choice by the official to remove the tax entirely.

**Why show the tax alert for any room in a governed zone rather than only tagged shop rooms?**
The requirement says "all shop listings in a zone should have text indicating the tax rate." The listing itself (via the mutated price) carries that information. The room alert is supplemental context. Showing it zone-wide for governed zones is simpler and more informative than requiring every shop room to carry an additional tag.

**Why not add a `ZoneName` field to `ShopListRequest` instead of `Room`?**
`Room` is more general and consistent with how other hooks (e.g., `OnRoomLook`) pass context. A future module may need other room properties (e.g., tags, room ID). Passing the full room pointer is the established pattern in this codebase.
