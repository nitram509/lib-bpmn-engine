package main

import (
	_ "embed"
	"net/http"
)

func handleIndex(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.Write(IndexHtml)
}

func handleOrderingItemsWorkflowBpmn(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/xml")
	writer.Write(OrderingItemsWorkflowBpmn)
}

func handleShowProcess(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.Write(ShowProcessHtml)
}

//go:embed ordering-items-workflow.bpmn
var OrderingItemsWorkflowBpmn []byte

//go:embed index.html
var IndexHtml []byte

//go:embed show-process.html
var ShowProcessHtml []byte
