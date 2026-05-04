package events

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["eventListeners"] = util.MemoryResult{Memory: util.MemoryUsage(eventListeners), Count: len(eventListeners)}
	ret["eventsWithoutListeners"] = util.MemoryResult{Memory: util.MemoryUsage(eventsWithoutListeners), Count: len(eventsWithoutListeners)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Events`, GetMemoryUsage)
}
