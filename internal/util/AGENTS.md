# GoMud Util System Context

## Overview

The GoMud util system provides essential utility functions and infrastructure services including game timing management, string processing, file operations, cryptographic functions, memory monitoring, performance tracking, and a generic synchronous hook mechanism. It serves as the foundational layer supporting all other game systems.

## Files

| File | Purpose |
|------|---------|
| `util.go` | Core utilities: timing, string processing, hashing, file I/O, dice, compression |
| `hook.go` | Generic type-safe synchronous callback chain (`Hook[T]`) |
| `memory.go` | Memory reporting, `MemoryUsage`, `ServerStats`, `FormatBytes` |
| `copyover.go` | Copyover (hot-restart) helpers |

## Game Timing System

### Turn and Round Counting

```go
const (
    RoundCountMinimum  = 1314000       // ~4 years offset for delta safety
    RoundCountFilename = ".roundcount" // Persistence file name
)

func IncrementTurnCount() uint64  // Increment and return the turn counter
func GetTurnCount() uint64        // Read the current turn counter
func IncrementRoundCount() uint64 // Increment and return the round counter
func GetRoundCount() uint64       // Read the current round counter
func SetRoundCount(newRoundCount uint64) // Override the round counter
```

Turns are the smallest time unit (driven by `TurnMs`). Rounds are coarser (driven by `RoundSeconds`). The round counter starts at `RoundCountMinimum` to keep delta comparisons safe even after a fresh start.

### Round Count Persistence

```go
func SaveRoundCount(fpath string)      // Write current roundCount to file
func LoadRoundCount(fpath string) uint64 // Read roundCount from file; creates file if missing
```

### High-Level Mutex

A single `sync.RWMutex` (`mudLock`) synchronizes high-level access to game data between asynchronous components.

```go
func LockMud()    // Exclusive write lock
func UnlockMud()  // Release write lock
func RLockMud()   // Shared read lock
func RUnlockMud() // Release read lock
```

### Performance Tracking (Accumulator)

```go
type Accumulator struct {
    Name    string
    Total   float64
    Lowest  float64
    Highest float64
    Count   float64
    Start   time.Time
}

func (t *Accumulator) Record(nextValue float64) // Add a sample
func (t *Accumulator) Average() float64         // Compute mean
func (t *Accumulator) Stats() (lowest, highest, average, count float64)

func TrackTime(name string, timePassed float64) // Record a timing sample by name
func GetTimeTrackers() []Accumulator            // Return all accumulated stats
```

## String Processing

### Text Matching

```go
// FindMatchIn searches a list of strings for the best match.
// Returns (exactMatch, closeMatch). Supports "#N" suffix for nth-match selection
// (e.g. "sword#2" finds the second sword). Falls back to substring search if no
// prefix/exact match is found.
func FindMatchIn(searchName string, items ...string) (match string, closeMatch string)

// GetMatchNumber splits "sword#2" -> ("sword", 2). Returns 1 if no # present.
func GetMatchNumber(input string) (string, int)

// BreakIntoParts splits "a b c" into ["a b c", "b c", "c"] for progressive matching.
func BreakIntoParts(full string) []string
```

### Preposition Stripping

```go
// StripPrepositions removes leading/embedded prepositions:
// onto, into, over, to, toward, towards, from, in, under, upon, with, the, my
func StripPrepositions(input string) string
```

### String Splitting

```go
// SplitString word-wraps text to lineWidth, respecting CJK character widths.
// Handles existing \n line breaks and punctuation attachment.
func SplitString(input string, lineWidth int) []string

// SplitStringNL is like SplitString but joins with \r\n (and optional line prefix).
func SplitStringNL(input string, lineWidth int, nlPrefix ...string) string

// SplitButRespectQuotes splits on whitespace but keeps quoted strings together.
// Removes surrounding quotes from matched tokens.
func SplitButRespectQuotes(s string) []string
```

### Wildcard Matching

```go
// StringWildcardMatch matches stringToSearch against a pattern.
// Supports leading * (endsWith), trailing * (startsWith), both * (contains).
func StringWildcardMatch(stringToSearch string, patternToSearch string) bool
```

### Color Code Processing

```go
// ConvertColorShortTags converts {fg:bg} short tags to <ansi> XML tags.
// Example: {1:2} -> <ansi fg="1" bg="2">
func ConvertColorShortTags(input string) string

// StripANSI removes ANSI escape sequences (\x1b[...m) from a string.
func StripANSI(str string) string

// StripCharsForScreenReaders replaces box-drawing and decorative characters
// with spaces for screen-reader accessibility.
func StripCharsForScreenReaders(s string) string
```

### Filename Sanitization

```go
// ConvertForFilename lowercases input and replaces any character that is not
// [a-z0-9] with '_'. Apostrophes are silently dropped.
func ConvertForFilename(input string) string

// FilePath joins path parts using the OS path separator (filepath.FromSlash).
func FilePath(pathParts ...string) string
```

### Number Formatting

```go
// FormatNumber formats an integer with comma separators (e.g. 1234567 -> "1,234,567").
func FormatNumber(n int) string

// BoolYN returns "yes" or "no".
func BoolYN(b bool) string
```

### Visual / UI Helpers

```go
// ProgressBar returns (fullBar, emptyBar) strings using block characters.
// complete is a 0.0–1.0 fraction. Custom bar characters can be provided.
func ProgressBar(complete float64, maxBarSize int, barParts ...string) (fullBar string, emptyBar string)

// HealthClass returns a CSS class string like "health-70" quantized to 10s.
func HealthClass(health int, maxHealth int) string

// ManaClass returns a CSS class string like "mana-50" quantized to 10s.
func ManaClass(mana int, maxMana int) string

// QuantizeTens returns value/max as a percentage quantized to the nearest 10.
func QuantizeTens(value int, max int) int

// PercentOfTotal returns (value1+value2)/value1 as a float64.
func PercentOfTotal(value1 int, value2 int) float64
```

## Cryptographic and Hashing

```go
// Hash returns the hex-encoded SHA-256 of input (used for legacy password hashing).
func Hash(input string) string

// HashBytes returns the hex-encoded SHA-256 of a byte slice.
func HashBytes(input []byte) string

// Md5 returns the hex-encoded MD5 of a string.
func Md5(input string) string

// Md5Bytes returns the raw MD5 bytes of a byte slice.
func Md5Bytes(input []byte) []byte
```

### Lock Sequence Generation

```go
// GetLockSequence generates a deterministic UP/DOWN sequence (e.g. "UUDUDD")
// from a lock identifier, difficulty (2–32 steps), and seed string.
// Used for puzzle locks in the game world.
func GetLockSequence(lockIdentifier string, difficulty int, seed string) string
```

## Random Number Generation

```go
// Rand returns a random int in [0, maxInt). Returns 0 if maxInt < 1.
func Rand(maxInt int) int

// RollDice rolls `dice` dice with `sides` sides each, summing the results.
// Negative dice count inverts the total. Uses Rand internally.
func RollDice(dice int, sides int) int

// LogRoll logs a debug message for a roll result vs target number.
func LogRoll(name string, rollResult int, targetNumber int)
```

### Dice Roll Parsing

```go
// ParseDiceRoll parses a dice expression like "2@1d6+3#9,11" into components.
// Format: [attacks@]dCount d dSides [+/-bonus] [#buffId,buffId,...]
func ParseDiceRoll(dRoll string) (attacks int, dCount int, dSides int, bonus int, buffOnCrit []int)

// FormatDiceRoll serializes the components back into a dice expression string.
func FormatDiceRoll(attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) string
```

## Data Compression and Encoding

```go
// Compress gzip-compresses a byte slice. Returns empty slice on error.
func Compress(input []byte) []byte

// Decompress gzip-decompresses a byte slice. Returns empty slice on error.
func Decompress(input []byte) []byte

// Encode base64-encodes a byte slice (standard encoding).
func Encode(blobdata []byte) string

// Decode base64-decodes a string. Ignores errors (returns nil on failure).
func Decode(base64str string) []byte
```

## File Operations

```go
// SafeSave writes data to path+".new" then renames it to path.
// Reduces risk of corruption from interrupted writes.
func SafeSave(path string, data []byte) error

// Save writes data to path. If doSafe is true, delegates to SafeSave.
func Save(path string, data []byte, doSafe ...bool) error

// ValidateWorldFiles checks that all subdirectories present in exampleWorldPath
// also exist in worldPath. Returns an error if any are missing.
func ValidateWorldFiles(exampleWorldPath string, worldPath string) error
```

## Network Utilities

```go
// SetServerAddress / GetServerAddress manage a global server address string.
func SetServerAddress(addr string)
func GetServerAddress() string

// GetMyIP fetches the server's public IP from api.ipify.org.
func GetMyIP() string
```

## Memory Management (`memory.go`)

```go
type MemResultUnit uint8

const (
    UnitBytes MemResultUnit = iota // Memory holds a byte count
    UnitCount                      // Memory holds a plain integer count (not bytes)
)

type MemReport func() map[string]MemoryResult

type MemoryResult struct {
    Memory uint64
    Count  int
    Unit   MemResultUnit
}

// AddMemoryReporter registers a named memory reporter function.
func AddMemoryReporter(name string, reporter MemReport)

// GetMemoryReport calls all registered reporters and returns their results.
func GetMemoryReport() (names []string, trackedResults []map[string]MemoryResult)

// ApproximateMemoryUsage estimates the memory footprint of any value using reflection.
// Handles pointers (with deduplication), slices (capacity-based), structs, maps,
// arrays, strings (header + backing array), and interface values.
// Returns 0 for nil.
func ApproximateMemoryUsage(i interface{}) uint64

// MemoryUsage is an alias for ApproximateMemoryUsage for backwards compatibility.
func MemoryUsage(i interface{}) uint64

// FormatBytes formats a byte count as a human-readable string (B, KB, MB, GB, etc.).
// Returns "0 B" for zero.
func FormatBytes(bytes uint64) string

// ServerStats returns a formatted ANSI string with Go runtime memory stats
// (heap, stack, total, GC count, GOMAXPROCS, goroutine count).
func ServerStats() string

// ServerGetMemoryUsage returns a MemoryResult map of Go runtime stats.
// Entries with Unit == UnitCount hold plain integer counts, not bytes.
// Automatically registered as the "Go" memory reporter on init.
func ServerGetMemoryUsage() map[string]MemoryResult
```

## Synchronous Hook System (`hook.go`)

`Hook[T]` is a generic, type-safe synchronous callback chain. Each registered handler receives the current value, may modify it, and returns the (possibly modified) value. All handlers run in registration order. The final value is returned to the caller.

```go
type Hook[T any] struct { ... }

// Register adds a handler. Thread-safe.
func (h *Hook[T]) Register(fn func(T) T)

// Fire runs all handlers in order, threading the return value through each.
// Thread-safe. Returns the original value unchanged if no handlers are registered.
func (h *Hook[T]) Fire(data T) T
```

### Defining a Hook Point

```go
// In the package that owns the data:
var OnGetDetails util.Hook[RoomTemplateDetails]

// At the call site:
details = rooms.OnGetDetails.Fire(details)
```

### Registering a Handler

```go
// From a module or plugin:
rooms.OnGetDetails.Register(func(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
    // modify d...
    return d
})
```

### Currently Defined Hook Points

- **`rooms.OnGetDetails`** (`util.Hook[RoomTemplateDetails]`) — fired at the end of `rooms.GetDetails`, allowing modules to modify room details before they reach the caller.

## Dependencies

- `crypto/sha256`, `crypto/md5` - Hashing
- `compress/gzip` - Data compression
- `encoding/base64` - Base64 encoding
- `math/rand` - Pseudo-random number generation
- `sync` - Mutex for `mudLock` and `Hook[T]`
- `reflect` - `MemoryUsage` implementation
- `runtime` - `ServerStats` / `ServerGetMemoryUsage`
- `path/filepath` - Cross-platform path handling
- `github.com/mattn/go-runewidth` - CJK character width for `SplitString`
- `internal/mudlog` - Debug logging for `LogRoll`, `LoadRoundCount`
- `internal/term` - `CRLFStr` used in `SplitStringNL` and `ServerStats`
