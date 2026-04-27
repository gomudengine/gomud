package colorpatterns

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["numericPatterns"] = util.MemoryResult{Memory: util.MemoryUsage(numericPatterns), Count: len(numericPatterns)}
	ret["ShortTagPatterns"] = util.MemoryResult{Memory: util.MemoryUsage(ShortTagPatterns), Count: len(ShortTagPatterns)}

	return ret
}

func init() {
	util.AddMemoryReporter(`ColorPatterns`, GetMemoryUsage)
}
