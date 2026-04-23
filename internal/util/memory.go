package util

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/term"
)

// MemResultUnit describes what the Memory field of a MemoryResult holds so
// callers can format it correctly.
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

type memReporter struct {
	name     string
	reporter MemReport
}

var (
	memoryReportersMu sync.RWMutex
	memoryReporters   []memReporter
)

// AddMemoryReporter registers a named memory reporter. Duplicate names are
// accepted but will appear as separate sections in the report.
func AddMemoryReporter(name string, reporter MemReport) {
	memoryReportersMu.Lock()
	defer memoryReportersMu.Unlock()
	memoryReporters = append(memoryReporters, memReporter{name: name, reporter: reporter})
}

// GetMemoryReport calls every registered reporter and returns the section
// names and their results. The two slices are parallel: names[i] corresponds
// to trackedResults[i].
func GetMemoryReport() (names []string, trackedResults []map[string]MemoryResult) {
	memoryReportersMu.RLock()
	reporters := make([]memReporter, len(memoryReporters))
	copy(reporters, memoryReporters)
	memoryReportersMu.RUnlock()

	names = make([]string, 0, len(reporters))
	trackedResults = make([]map[string]MemoryResult, 0, len(reporters))

	for _, r := range reporters {
		names = append(names, r.name)
		trackedResults = append(trackedResults, r.reporter())
	}

	return names, trackedResults
}

// ServerStats returns a formatted ANSI string summarising Go runtime memory
// and scheduler statistics. A single ReadMemStats call is shared with
// ServerGetMemoryUsage when both are needed; call ServerGetMemoryUsageFromStats
// directly to reuse an already-read MemStats value.
func ServerStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return serverStatsFromMemStats(&m)
}

func serverStatsFromMemStats(m *runtime.MemStats) string {
	return fmt.Sprintf(
		`<ansi fg="yellow-bold">Heap:</ansi>       <ansi fg="green-bold">%dMB</ansi> <ansi fg="yellow-bold">Largest Heap:</ansi>  <ansi fg="green-bold">%dMB</ansi>`+term.CRLFStr+
			`<ansi fg="yellow-bold">Stack:</ansi>      <ansi fg="green-bold">%dMB</ansi>`+term.CRLFStr+
			`<ansi fg="yellow-bold">Total Mem:</ansi>  <ansi fg="green-bold">%dMB</ansi>`+term.CRLFStr+
			`<ansi fg="yellow-bold">GC ct:</ansi>      <ansi fg="green-bold">%d</ansi>`+term.CRLFStr+
			`<ansi fg="yellow-bold">NumCPU:</ansi>     <ansi fg="green-bold">%d</ansi>`+term.CRLFStr+
			`<ansi fg="yellow-bold">Goroutines:</ansi> <ansi fg="green-bold">%d</ansi>`,
		m.HeapAlloc/1024/1024, m.HeapSys/1024/1024,
		m.StackSys/1024/1024,
		m.Sys/1024/1024,
		m.NumGC,
		runtime.GOMAXPROCS(0),
		runtime.NumGoroutine())
}

// ServerGetMemoryUsage reads runtime stats and returns them as a MemoryResult
// map. Entries whose Unit is UnitCount hold plain integer counts, not bytes.
func ServerGetMemoryUsage() map[string]MemoryResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return serverGetMemoryUsageFromStats(&m)
}

func serverGetMemoryUsageFromStats(m *runtime.MemStats) map[string]MemoryResult {
	ret := map[string]MemoryResult{}
	ret[`HeapAlloc (!Freed)`] = MemoryResult{Memory: m.HeapAlloc, Unit: UnitBytes}
	ret[`HeapSys (!Reclaimed)`] = MemoryResult{Memory: m.HeapSys, Unit: UnitBytes}
	ret[`StackSys (Reserved)`] = MemoryResult{Memory: m.StackSys, Unit: UnitBytes}
	ret[`StackInuse (In Use)`] = MemoryResult{Memory: m.StackInuse, Unit: UnitBytes}
	ret[`Sys (Everything)`] = MemoryResult{Memory: m.Sys, Unit: UnitBytes}
	ret[`GC Count`] = MemoryResult{Memory: uint64(m.NumGC), Unit: UnitCount}
	ret[`Maximum Processors`] = MemoryResult{Memory: uint64(runtime.GOMAXPROCS(0)), Unit: UnitCount}
	ret[`Goroutines Count`] = MemoryResult{Memory: uint64(runtime.NumGoroutine()), Unit: UnitCount}
	return ret
}

// sizeOf recursively estimates the memory occupied by v. It uses slice
// capacity (not just length), counts string backing arrays, unwraps interface
// values, and deduplicates pointer targets via the seen map.
func sizeOf(v reflect.Value, seen map[uintptr]struct{}) uintptr {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}
		ptr := v.Pointer()
		if _, visited := seen[ptr]; visited {
			return 0
		}
		seen[ptr] = struct{}{}
		return v.Type().Size() + sizeOf(v.Elem(), seen)

	case reflect.Slice:
		if v.IsNil() {
			return 0
		}
		// Use capacity, not length, to count allocated backing memory.
		cap_ := v.Cap()
		if cap_ == 0 {
			return 0
		}
		elemSize := sizeOf(reflect.New(v.Type().Elem()).Elem(), seen)
		return uintptr(cap_) * elemSize

	case reflect.Struct:
		var size uintptr
		for i := 0; i < v.NumField(); i++ {
			size += sizeOf(v.Field(i), seen)
		}
		return size

	case reflect.Array:
		length := v.Len()
		if length == 0 {
			return 0
		}
		elemSize := sizeOf(reflect.New(v.Type().Elem()).Elem(), seen)
		return uintptr(length) * elemSize

	case reflect.String:
		// string header (pointer + length) plus the backing byte array.
		return v.Type().Size() + uintptr(v.Len())

	case reflect.Map:
		if v.IsNil() {
			return 0
		}
		var size uintptr
		for _, key := range v.MapKeys() {
			size += sizeOf(key, seen) + sizeOf(v.MapIndex(key), seen)
		}
		return size

	case reflect.Interface:
		if v.IsNil() {
			return 0
		}
		return sizeOf(v.Elem(), seen)

	default:
		return v.Type().Size()
	}
}

// ApproximateMemoryUsage returns a rough estimate of the memory occupied by i.
// It accounts for slice capacity, string backing arrays, interface values, and
// deduplicates shared pointer targets. Map internal bucket overhead is not
// counted. Returns 0 for nil.
func ApproximateMemoryUsage(i interface{}) uint64 {
	if i == nil {
		return 0
	}
	return uint64(sizeOf(reflect.ValueOf(i), make(map[uintptr]struct{})))
}

// MemoryUsage is an alias for ApproximateMemoryUsage retained for backwards
// compatibility with existing callers.
func MemoryUsage(i interface{}) uint64 {
	return ApproximateMemoryUsage(i)
}

// FormatBytes formats a byte count as a human-readable string.
// Returns "0 B" for zero, otherwise scales to KB/MB/GB/TB/PB/EB.
func FormatBytes(bytes uint64) string {
	if bytes == 0 {
		return "    0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%5.1f  B", float64(bytes))
	}

	exp := int(math.Log(float64(bytes)) / math.Log(unit))
	prefixes := "KMGTPE"
	if exp-1 >= len(prefixes) {
		exp = len(prefixes)
	}
	prefix := prefixes[exp-1]
	return fmt.Sprintf("%5.1f %cB", float64(bytes)/math.Pow(unit, float64(exp)), prefix)
}

func init() {
	AddMemoryReporter(`Go`, ServerGetMemoryUsage)
}
