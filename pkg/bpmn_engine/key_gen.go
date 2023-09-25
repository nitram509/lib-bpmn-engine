package bpmn_engine

import (
	"hash/adler32"
	"os"

	"github.com/bwmarrin/snowflake"
)

var globalIdGenerator *snowflake.Node = nil

func (state *BpmnEngineState) generateKey() int64 {
	return state.snowflake.Generate().Int64()
}

// getGlobalSnowflakeIdGenerator the global ID generator
// constraints: see also createGlobalSnowflakeIdGenerator
func getGlobalSnowflakeIdGenerator() *snowflake.Node {
	if globalIdGenerator == nil {
		globalIdGenerator = createGlobalSnowflakeIdGenerator()
	}
	return globalIdGenerator
}

// createGlobalSnowflakeIdGenerator a new ID generator,
// constraints: creating two new instances within a few microseconds, will create generators with the same seed
func createGlobalSnowflakeIdGenerator() *snowflake.Node {
	hash32 := adler32.New()
	for _, e := range os.Environ() {
		hash32.Sum([]byte(e))
	}
	snowflakeNode, err := snowflake.NewNode(int64(hash32.Sum32()))
	if err != nil {
		panic("Can't initialize snowflake ID generator. Message: " + err.Error())
	}
	return snowflakeNode
}
