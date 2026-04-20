# Gambling Module Context

## Overview

The `gambling` module is a self-contained plugin that adds gambling fixtures to the game world
without any changes to the core engine. It provides gambling items with interactive scripts,
a slot machine, and a claw machine. Rooms opt into each fixture via room tags.

---

## Room Tags

Rooms can opt into gambling features by adding specific tags. Tags are set via the
admin `room tag` command or by editing the room's YAML file directly:

```yaml
tags:
  - slot machine
```

- A single room may carry both tags and have both machines present.
- Tags are case-insensitive at the module level (`Slot Machine`, `slot machine`, and
  `SLOT MACHINE` are all treated the same).
- Machines are purely virtual; no item or mob needs to be placed in the room.

---

## Slot Machine

**Tags:** `slots` or `slot machine` (either is accepted)

Adds a slot machine to the room. Players interact with it using:

- `play slots` - spend gold to spin the reels
- `look slot machine` - view cost, current jackpot, and payout table

**Configuration keys** (`Modules.gambling.*`):

| Key | Default | Current | Description |
|---|---|---|---|
| `SlotCost` | `10` | `25` | Gold cost per spin |

Half of every wager feeds the shared jackpot pool (`cost / 2` = 12 gold per spin at current cost). The jackpot persists across restarts.

### Symbol Table

Each reel has a total weight of **100**, so each symbol's weight is directly its per-reel
percentage probability.

| Symbol | Weight | Per-reel probability |
|--------|--------|----------------------|
| cherry | 30     | 30.00%               |
| lemon  | 25     | 25.00%               |
| orange | 20     | 20.00%               |
| plum   | 15     | 15.00%               |
| bell   |  7     |  7.00%               |
| bar    |  2     |  2.00%               |
| seven  |  1     |  1.00%               |

All three reels are spun independently.

### Outcome Evaluation Order

`evaluate(a, b, c)` checks conditions in this exact order:

1. All three identical → **triple** (special cases: seven = JACKPOT, bar = TRIPLE BAR 20x, bell = TRIPLE BELL 10x, all others = TRIPLE \<SYMBOL\> 5x)
2. Any two identical → **PAIR** (2x)
3. Two or more cherries → **CHERRIES** (2x) — **this branch is unreachable** (see note below)
4. Default → **loss** (0x)

> **Note on the CHERRIES branch:** Any combination of exactly two cherries is caught by the
> PAIR check at step 2 before step 3 is reached. Triple cherry is caught at step 1. The
> CHERRIES branch therefore can never fire with the current logic. It is dead code.

### Probability and Combination Table

Total possible weighted combinations: **100^3 = 1,000,000**

| Outcome        | Payout | Combinations | Probability  | Approx. odds |
|----------------|--------|-------------|--------------|--------------|
| JACKPOT (777)  | jackpot pool | 1       | 0.0001%      | 1 in 1,000,000 |
| TRIPLE BAR     | 20x    | 8           | 0.0008%      | 1 in 125,000 |
| TRIPLE BELL    | 10x    | 343         | 0.0343%      | 1 in ~2,915  |
| TRIPLE CHERRY  | 5x     | 27,000      | 2.7000%      | 1 in ~37     |
| TRIPLE LEMON   | 5x     | 15,625      | 1.5625%      | 1 in 64      |
| TRIPLE ORANGE  | 5x     | 8,000       | 0.8000%      | 1 in 125     |
| TRIPLE PLUM    | 5x     | 3,375       | 0.3375%      | 1 in ~296    |
| *All triples*  | *varies* | *54,352*  | *5.4352%*    | *1 in ~18*   |
| PAIR (any)     | 2x     | 498,144     | 49.8144%     | ~1 in 2      |
| CHERRIES       | 2x     | 0           | 0% (unreachable) | —        |
| Loss           | 0x     | 447,504     | 44.7504%     | ~1 in 2.2    |

Combination counts are computed as weighted counts out of 1,000,000 (since total weight = 100
per reel, each weight unit equals one combination unit per reel).

Pair combinations = 3 × Σ(w² × (100 − w)) over all symbols = 498,144.

### Expected Return (Fixed Payouts Only)

Excluding the jackpot (which is variable), the expected gold returned per 1 gold wagered from
fixed-payout outcomes is:

```
E = (27000×5 + 15625×5 + 8000×5 + 3375×5 + 343×10 + 8×20 + 498144×2) / 1,000,000
  = (135000 + 78125 + 40000 + 16875 + 3430 + 160 + 996288) / 1,000,000
  = 1,269,878 / 1,000,000
  ≈ 1.2699
```

**Fixed-payout RTP: ~127%** — the fixed payouts alone return more than the wager.

### Jackpot Economics

Half of every wager (`cost / 2`) is added to the shared jackpot pool, currently 12 gold per spin. The jackpot is won
when three sevens land (probability 1 in 1,000,000 per spin). The pool resets to zero on
a jackpot win and accumulates again from subsequent plays.

Because the fixed-payout RTP already exceeds 100%, the jackpot pool effectively represents
a bonus on top of already-positive expected returns for players.

---

## Claw Machine

**Tag:** `claw machine`

Adds a claw machine to the room. Players interact with it using:

- `play claw machine` - spend gold for one claw attempt
- `look claw machine` - view cost, win chance, and prize list

**Configuration keys** (`Modules.gambling.*`):

| Key | Default | Current | Description |
|---|---|---|---|
| `ClawCost` | `10` | `100` | Gold cost per attempt |
| `ClawWinChance` | `10` | `10` | Percentage chance (1-100) of winning a prize |
| `ClawPrizeDie` | `20` | `30` | Relative weight for the 6-sided die |
| `ClawPrizeCoin` | `20` | `2` | Relative weight for the lucky coin |
| `ClawPrizeBottle` | `20` | `8` | Relative weight for the empty bottle |
| `ClawPrizeCards` | `15` | `20` | Relative weight for the deck of cards |
| `ClawPrize8Ball` | `15` | `30` | Relative weight for the magic 8-ball |
| `ClawPrizeTarot` | `10` | `10` | Relative weight for the tarot deck |

Weights are relative to each other. The current weights sum to 100, so each weight is directly
the percentage chance of that prize being selected on a win. Set a weight to `0` to remove
that prize from the pool entirely. The `look claw machine` command shows each active prize
with its calculated percentage chance based on current weights.

**Current prize distribution (on a win):**

| Prize | Weight | Chance on win |
|---|---|---|
| 6-sided die | 30 | 30% |
| magic 8-ball | 30 | 30% |
| deck of cards | 20 | 20% |
| tarot deck | 10 | 10% |
| empty bottle | 8 | 8% |
| lucky coin | 2 | 2% |
