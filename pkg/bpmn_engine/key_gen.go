package bpmn_engine

import (
	"github.com/bwmarrin/snowflake"
	"hash/adler32"
	"os"
)

func (state *BpmnEngineState) generateKey() int64 {
	return state.snowflake.Generate().Int64()
}

func initializeSnowflakeIdGenerator() *snowflake.Node {
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
