package buffs

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["buffs"] = util.MemoryResult{Memory: util.MemoryUsage(buffs), Count: len(buffs)}
	ret["buffflags"] = util.MemoryResult{Memory: util.MemoryUsage(flagSpecs), Count: len(flagSpecs)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Buffs`, GetMemoryUsage)
}
