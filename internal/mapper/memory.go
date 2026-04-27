package mapper

import (
	"sync/atomic"

	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	pathCacheHits   atomic.Uint64
	pathCacheMisses atomic.Uint64
)

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["mapperZoneCache"] = util.MemoryResult{Memory: util.MemoryUsage(mapperZoneCache), Count: len(mapperZoneCache)}
	ret["roomIdToMapperCache"] = util.MemoryResult{Memory: util.MemoryUsage(roomIdToMapperCache), Count: len(roomIdToMapperCache)}
	ret["pathCache"] = util.MemoryResult{Memory: util.MemoryUsage(pathCache), Count: pathCache.Len()}
	ret["pathCacheHits"] = util.MemoryResult{Memory: pathCacheHits.Load(), Unit: util.UnitCount}
	ret["pathCacheMisses"] = util.MemoryResult{Memory: pathCacheMisses.Load(), Unit: util.UnitCount}

	return ret
}

func init() {
	util.AddMemoryReporter(`Mapper`, GetMemoryUsage)
}
