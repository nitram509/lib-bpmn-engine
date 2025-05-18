package main

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter/zeebe"
	"github.com/redis/go-redis/v9"
)

func main() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New()
	// you can use hazelcast or redis as exporter
	// hazelcastExporter, err := hazelcastExporter()
	// if err != nil {
	// 	panic("hazelcast exporter can't be created.")
	// }
	redisExporter, err := redisExporter()
	if err != nil {
		panic("redis exporter can't be created.")
	}
	// register the exporter
	bpmnEngine.AddEventExporter(redisExporter)
	// basic example loading a BPMN from file,
	process, err := bpmnEngine.LoadFromFile("simple_task.bpmn")
	if err != nil {
		panic("file \"simple_task.bpmn\" can't be read.")
	}
	// register a handler for a service task by defined task type
	bpmnEngine.NewTaskHandler().Id("hello-world").Handler(printContextHandler)
	// and execute the process
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	println(fmt.Sprintf("instanceKey=%d", instance.GetInstanceKey()))
}

func printContextHandler(job bpmn_engine.ActivatedJob) {
	// trivial handler is requires
	job.Complete()
}

func redisExporter() (exporter.EventExporter, error) {
	// the exporter will require a running Redis server at localhost:6379
	redisExporter, err := zeebe.NewRedisExporter(redis.Options{
		Addr:    "localhost:6379",
		Network: "tcp",
	})
	if err != nil {
		return nil, err
	}
	return &redisExporter, nil
}

func hazelcastExporter() (exporter.EventExporter, error) {
	// the exporter will require a running Hazelcast cluster at 127.0.0.1:5701
	ctx := context.TODO()
	config := hazelcast.Config{}
	config.Cluster.Network.SetAddresses("localhost:5701")
	client, err := hazelcast.StartNewClientWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	// create the exporter
	exporter, err := zeebe.NewExporterWithHazelcastClient(client)
	if err != nil {
		return nil, err
	}
	return &exporter, nil
}
