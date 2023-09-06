package main

import (
	_ "embed"
	"net/http"
	"strconv"
)

func handleReceivePayment(writer http.ResponseWriter, request *http.Request) {
	orderIdStr := request.FormValue("orderId")
	amount := request.FormValue("amount")
	if len(orderIdStr) > 0 && len(amount) > 0 {
		orderId, _ := strconv.ParseInt(orderIdStr, 10, 64)
		processInstance := bpmnEngine.FindProcessInstance(orderId)
		if processInstance != nil {
			vars := map[string]interface{}{
				"amount": amount,
			}
			bpmnEngine.PublishEventForInstance(processInstance.GetInstanceKey(), "payment-received", vars)
			bpmnEngine.RunOrContinueInstance(processInstance.GetInstanceKey())
			http.Redirect(writer, request, "/", http.StatusFound)
			return
		}
	}
	writer.WriteHeader(400)
	writer.Write([]byte("Bad request: the request must contain form data with 'orderId' and 'amount', and the order must exist"))
}
