package pets

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["petTypes"] = util.MemoryResult{Memory: util.MemoryUsage(petTypes), Count: len(petTypes)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Pets`, GetMemoryUsage)
}
