package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
)

func handleOrder(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		createNewOrder(writer, request)
	} else if request.Method == "GET" {
		showOrderStatus(writer, request)
	}
}

func createNewOrder(writer http.ResponseWriter, request *http.Request) {
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	redirectUrl := fmt.Sprintf("/show-process.html?orderId=%d", instance.GetInstanceKey())
	http.Redirect(writer, request, redirectUrl, http.StatusFound)
}

func showOrderStatus(writer http.ResponseWriter, request *http.Request) {
	orderIdStr := request.URL.Query()["orderId"][0]
	orderId, _ := strconv.ParseInt(orderIdStr, 10, 64)
	instance := bpmnEngine.FindProcessInstanceById(orderId)
	if instance != nil {
		// we re-use this GET request to ensure we catch up the timers - ideally the service uses internal timers instead
		bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
		bytes, _ := prepareJsonResponse(orderIdStr, instance.GetState(), instance.GetCreatedAt())
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
		return
	}
	http.NotFound(writer, request)
}
