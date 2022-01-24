package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"net/http"
	"strconv"
	"time"
)

var bpmnEngine bpmn_engine.BpmnEngineState
var process *bpmn_engine.ProcessInfo

// main does start a trivial microservice, listening on port 8080
// open your web browser with http://localhost:8080/
func main() {
	initBpmnEngine()
	http.HandleFunc("/api/order", handleOrder)                                        // POST new Order
	http.HandleFunc("/api/receive-payment", handleReceivePayment)                     // webhook from the payment system
	http.HandleFunc("/show-process.html", handleShowProcess)                          // the index page
	http.HandleFunc("/index.html", handleIndex)                                       // the index page
	http.HandleFunc("/", handleIndex)                                                 // the index page
	http.HandleFunc("/ordering-items-workflow.bpmn", handleOrderingItemsWorkflowBpmn) // the BPMN file, for documentation purpose
	http.ListenAndServe(":8080", nil)
}

func initBpmnEngine() {
	bpmnEngine = bpmn_engine.New("Ordering-Microservice")
	process, _ = bpmnEngine.LoadFromBytes(OrderingItemsWorkflowBpmn)
	bpmnEngine.AddTaskHandler("validate-order", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-bill", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-friendly-reminder", businessActionHandler)
	bpmnEngine.AddTaskHandler("update-accounting", businessActionHandler)
	bpmnEngine.AddTaskHandler("package-and-deliver", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-cancellation", businessActionHandler)
}

func businessActionHandler(job bpmn_engine.ActivatedJob) {
	// do important stuff here
	msg := fmt.Sprintf("%s >>> Executing job '%s", time.Now(), job.ElementId)
	println(msg)
}

func handleOrder(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		createNewOrder(writer, request)
	} else if request.Method == "GET" {
		showOrderStatus(writer, request)
	}
}

func showOrderStatus(writer http.ResponseWriter, request *http.Request) {
	orderIdStr := request.URL.Query()["orderId"][0]
	orderId, _ := strconv.ParseInt(orderIdStr, 10, 64)
	instance := bpmnEngine.FindProcessInstanceById(orderId)
	if instance != nil {
		// prepare response
		type Order struct {
			OrderId              string    `json:"orderId"`
			ProcessInstanceState string    `json:"state"`
			CreatedAt            time.Time `json:"createdAt"`
		}
		order := Order{
			OrderId:              orderIdStr,
			ProcessInstanceState: string(instance.GetState()),
			CreatedAt:            instance.GetCreatedAt(),
		}
		bytes, _ := json.Marshal(order)
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
		return
	}
	http.NotFound(writer, request)
}

func createNewOrder(writer http.ResponseWriter, request *http.Request) {
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	redirectUrl := fmt.Sprintf("/show-process.html?orderId=%d", instance.GetInstanceKey())
	http.Redirect(writer, request, redirectUrl, http.StatusFound)
}

func handleReceivePayment(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
}
