package connections

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	lock.RLock()
	defer lock.RUnlock()

	ret := map[string]util.MemoryResult{}

	ret["netConnections"] = util.MemoryResult{Memory: util.MemoryUsage(netConnections), Count: len(netConnections)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Connections`, GetMemoryUsage)
}
