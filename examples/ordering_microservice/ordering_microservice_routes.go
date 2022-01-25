package main

import "net/http"

func initHttpRoutes() {
	http.HandleFunc("/api/order", handleOrder)                                        // POST new Order
	http.HandleFunc("/api/receive-payment", handleReceivePayment)                     // webhook from the payment system
	http.HandleFunc("/show-process.html", handleShowProcess)                          // the index page
	http.HandleFunc("/index.html", handleIndex)                                       // the index page
	http.HandleFunc("/", handleIndex)                                                 // the index page
	http.HandleFunc("/ordering-items-workflow.bpmn", handleOrderingItemsWorkflowBpmn) // the BPMN file, for documentation purpose
}
