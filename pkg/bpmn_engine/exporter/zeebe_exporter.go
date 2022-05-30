package exporter

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"google.golang.org/protobuf/proto"
)

// hint:
// protoc --go_opt=paths=source_relative --go_out=. --go_opt=Mschema.proto=exporter/  schema.proto

type ZeebeExporter struct {
}

func (*ZeebeExporter) foo() {
	resources := make([]*DeploymentRecord_Resource, 1)
	resources[0] = &DeploymentRecord_Resource{
		ResourceName: "name",
		Resource:     []byte{},
	}
	resource := DeploymentRecord{
		Metadata: &RecordMetadata{
			Key:         123,
			Position:    456,
			PartitionId: 789,
			RecordType:  RecordMetadata_EVENT,
		},
		Resources: resources,
	}
	data, err := proto.Marshal(&resource)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}
	println(data)

	sendHazelcast(data)
}

func sendHazelcast(data []byte) {
	ctx := context.Background()
	// Start the client with defaults.
	client, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		panic(err)
	}
	// Get a reference to the queue.
	myQueue, err := client.GetRingBuffer(ctx, "my-queue")
	if err != nil {
		panic(err)
	}
	// Add an item to the queue if space is available (non-blocking).
	added, err := myQueue.Add(ctx, "item 1")
	if err != nil {
		panic(err)
	}
	fmt.Println("Added?", added)
	// Get the head of the queue if available and print item.
	item, err := myQueue.Poll(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Got item:", item)
}
