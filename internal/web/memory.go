package web

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["authCache"] = util.MemoryResult{Memory: util.MemoryUsage(authCache), Count: len(authCache)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Web`, GetMemoryUsage)
}
