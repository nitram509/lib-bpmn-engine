package main

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter/zeebe"
)

func main() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New()
	// the exporter will require a running Hazelcast cluster at 127.0.0.1:5701
	ctx := context.TODO()
	config := hazelcast.Config{}
	config.Cluster.Network.SetAddresses("localhost:5701")
	client, err := hazelcast.StartNewClientWithConfig(ctx, config)
	// create the client
	exporter, _ := zeebe.NewExporterWithHazelcastClient(client)
	// register the exporter
	bpmnEngine.AddEventExporter(&exporter)
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
