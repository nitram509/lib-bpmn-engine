package bpmn_engine

import (
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/labstack/gommon/log"
)

func TestSnowflakeParser(t *testing.T) {
	generator := createGlobalSnowflakeIdGenerator(102)
	log.Printf("Snowflake: %d", generator.Generate().Int64())

	// id := generator.Generate().Int64()

	node := ParseSnowflake(snowflake.ID(1852008629847195648))

	then.AssertThat(t, node.Node, is.EqualTo(int64(1)))
	then.AssertThat(t, node.Partition, is.EqualTo(int64(2)))
}
