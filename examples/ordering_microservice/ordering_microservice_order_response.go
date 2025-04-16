package main

import (
	_ "embed"
	"encoding/json"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"time"
)

func prepareJsonResponse(orderIdStr string, state bpmn_engine.ActivityState, createdAt time.Time) ([]byte, error) {
	type Order struct {
		OrderId              string    `json:"orderId"`
		ProcessInstanceState string    `json:"state"`
		CreatedAt            time.Time `json:"createdAt"`
	}
	order := Order{
		OrderId:              orderIdStr,
		ProcessInstanceState: string(state),
		CreatedAt:            createdAt,
	}
	return json.Marshal(order)
}
