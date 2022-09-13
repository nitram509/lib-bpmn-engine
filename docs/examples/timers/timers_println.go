package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func printScheduledTimerInformation(timer bpmn_engine.Timer) {
	println("State     : " + timer.State)
	println("CreatedAt : " + timer.CreatedAt.String())
	println("Duration  : " + timer.Duration.String())
	println("DueAt     : " + timer.DueAt.String())
}
