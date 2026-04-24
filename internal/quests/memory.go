package quests

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["quests"] = util.MemoryResult{Memory: util.MemoryUsage(quests), Count: len(quests)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Quests`, GetMemoryUsage)
}
