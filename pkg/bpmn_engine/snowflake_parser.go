package bpmn_engine

import (
	"github.com/bwmarrin/snowflake"
)

type SnowflakeNodeIdParsed struct {
	Node      int64
	Partition int64
}

func ParseSnowflake(id snowflake.ID) SnowflakeNodeIdParsed {
	nodeId := id.Node()

	node := nodeId / 100
	paritition := nodeId % 100

	return SnowflakeNodeIdParsed{
		Node:      node,
		Partition: paritition,
	}
}
