package spells

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["allSpells"] = util.MemoryResult{Memory: util.MemoryUsage(allSpells), Count: len(allSpells)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Spells`, GetMemoryUsage)
}
