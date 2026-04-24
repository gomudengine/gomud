package templates

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["templateCache"] = util.MemoryResult{Memory: util.MemoryUsage(templateCache), Count: len(templateCache)}
	ret["templateConfigCache"] = util.MemoryResult{Memory: util.MemoryUsage(templateConfigCache), Count: len(templateConfigCache)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Templates`, GetMemoryUsage)
}
