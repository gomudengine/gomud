package conversations

import "github.com/GoMudEngine/GoMud/internal/util"

func GetMemoryUsage() map[string]util.MemoryResult {
	ret := map[string]util.MemoryResult{}

	ret["converseCheckCache"] = util.MemoryResult{Memory: util.MemoryUsage(converseCheckCache), Count: len(converseCheckCache)}
	ret["conversations"] = util.MemoryResult{Memory: util.MemoryUsage(conversations), Count: len(conversations)}
	ret["conversationCounter"] = util.MemoryResult{Memory: util.MemoryUsage(conversationCounter), Count: len(conversationCounter)}

	return ret
}

func init() {
	util.AddMemoryReporter(`Conversations`, GetMemoryUsage)
}
