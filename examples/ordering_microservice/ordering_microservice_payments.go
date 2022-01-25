package main

import (
	_ "embed"
	"net/http"
)

func handleReceivePayment(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
}
