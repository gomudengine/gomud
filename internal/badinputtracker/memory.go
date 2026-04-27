package badinputtracker

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	lock.Lock()
	defer lock.Unlock()

	ret := map[string]util.MemoryResult{}

	ret["badCommands"] = util.MemoryResult{Memory: util.MemoryUsage(badCommands), Count: len(badCommands)}

	return ret
}

func init() {
	util.AddMemoryReporter(`BadInputTracker`, GetMemoryUsage)
}
