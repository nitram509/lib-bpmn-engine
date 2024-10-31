package bpmn_engine

import (
	"hash/adler32"
	"log"
	"os"

	"github.com/bwmarrin/snowflake"
)

var globalIdGenerator *snowflake.Node = nil

func (state *BpmnEngineState) generateKey() int64 {
	key := state.snowflake.Generate().Int64()
	node := ParseSnowflake(snowflake.ID(key))

	log.Printf("Generated key: %d parsed back %v", key, node)

	return key
}

// getGlobalSnowflakeIdGenerator the global ID generator
// constraints: see also createGlobalSnowflakeIdGenerator
func getGlobalSnowflakeIdGenerator(nodeId int) *snowflake.Node {
	log.Printf("Setting up snowflake generator with node ID: %d", nodeId)
	return createGlobalSnowflakeIdGenerator(nodeId)
}

// createGlobalSnowflakeIdGenerator a new ID generator,
// constraints: creating two new instances within a few microseconds, will create generators with the same seed
func createGlobalSnowflakeIdGenerator(nodeId int) *snowflake.Node {
	hash32 := adler32.New()
	for _, e := range os.Environ() {
		hash32.Sum([]byte(e))
	}
	snowflakeNode, err := snowflake.NewNode(int64(nodeId))
	if err != nil {
		panic("can't initialize snowflake ID generator. Message: " + err.Error())
	}
	return snowflakeNode
}
