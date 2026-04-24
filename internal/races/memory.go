package races

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["races"] = util.MemoryResult{Memory: util.MemoryUsage(races), Count: len(races)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Races`, GetMemoryUsage)
}
