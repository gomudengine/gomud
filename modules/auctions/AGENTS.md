# Auctions Module Context

## Overview

The `modules/auctions` module adds a global player-driven auction system to GoMud. Players can put items up for auction, other players bid on them in real time, and the highest bidder wins the item when the auction timer expires. Auction history is persisted across server restarts. The module also fires `AuctionUpdate` events consumed by the Discord integration.

## Key Components

### Module (`auctions.go`)

- Registered as plugin `auctions` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `auction`.
- Registers `OnLoad`/`OnSave` callbacks for persistence via `plug.ReadIntoStruct` / `plug.WriteStruct` (key: `auctionhistory`).
- Registers a `NewRound` listener for auction tick processing (countdown, reminders, expiry).

### Data Structures

```go
type AuctionsModule struct {
    plug       *plugins.Plugin
    auctionMgr AuctionManager
}

type AuctionManager struct {
    ActiveAuction   *ActiveAuctionItem
    maxHistoryItems int
    PastAuctions    []PastAuctionItem
}
```

### Command: `auction`

Subcommands:
- `auction` — displays the current auction status (if any).
- `auction <item> [min-bid] [anonymous]` — starts a new auction with the item from the player's inventory. Optional minimum bid and anonymous flag.
- `auction bid <amount>` — places a bid on the current auction.
- `auction history` — displays a table of past auctions.

Auction opt-out: players with `auction` config option set to `false` do not receive auction broadcasts.

### `AuctionUpdate` Event

The module fires `AuctionUpdate` (implements `events.GenericEvent`) with `State` values:
- `START` — new auction started
- `REMINDER` — periodic reminder during active auction
- `BID` — a bid was placed
- `END` — auction ended (with or without a winner)

This event is consumed by `internal/integrations/discord` for Discord notifications.

### Persistence

Auction history (up to `maxHistoryItems` = 10 past auctions) is saved to the plugin data store. Active auction state is also persisted, allowing in-progress auctions to survive server restarts.

### Configuration

From `Modules.auctions.*` in config:

| Key | Default | Description |
|---|---|---|
| `Duration` | (configured) | Auction duration in rounds |
| `MinimumBid` | (configured) | Default minimum bid if not specified by seller |

## File Structure

```
modules/auctions/
  auctions.go
  files/
    data-overlays/
      ansi-aliases.yaml
      config.yaml
      keywords.yaml
    datafiles/templates/auctions/
      auction-bid.template
      auction-end.template
      auction-start.template
      auction-update.template
    datafiles/templates/help/
      auction.template
```

## Dependencies

- `internal/events`: `NewRound` for auction ticking; `AuctionUpdate` event dispatch; `EquipmentChange` for gold tracking
- `internal/items`: Item lookup and transfer
- `internal/plugins`: Plugin and command registration
- `internal/rooms`, `internal/users`: Command execution context
- `internal/templates`: Auction message rendering
- `internal/util`: Argument parsing, number formatting
