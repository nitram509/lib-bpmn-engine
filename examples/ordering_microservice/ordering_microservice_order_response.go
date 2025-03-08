package main

import (
	_ "embed"
	"encoding/json"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"time"
)

func prepareJsonResponse(orderIdStr string, state BPMN20.ActivityState, createdAt time.Time) ([]byte, error) {
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
