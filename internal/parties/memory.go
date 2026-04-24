package parties

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["partyMap"] = util.MemoryResult{Memory: util.MemoryUsage(partyMap), Count: len(partyMap)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Parties`, GetMemoryUsage)
}
