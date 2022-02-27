package main

import "net/http"

func initHttpRoutes() {
	http.HandleFunc("/api/order", handleOrder)                                        // POST new or GET existing Order
	http.HandleFunc("/api/receive-payment", handleReceivePayment)                     // webhook for the payment system
	http.HandleFunc("/show-process.html", handleShowProcess)                          // shows the BPMN diagram
	http.HandleFunc("/index.html", handleIndex)                                       // the index page
	http.HandleFunc("/", handleIndex)                                                 // the index page
	http.HandleFunc("/ordering-items-workflow.bpmn", handleOrderingItemsWorkflowBpmn) // the BPMN file, for documentation purpose
}
