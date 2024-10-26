package server

import (
	"bytes"
	"compress/flate"
	"encoding/ascii85"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api"
)

type server struct {
	sync.RWMutex
	engine *bpmn_engine.BpmnEngineState
	port   int
}

func NewServer(engine *bpmn_engine.BpmnEngineState, port int) *server {
	return &server{
		engine: engine,
		port:   port,
	}
}

// Ensure that we implement the server interface
var _ api.ServerInterface = (*server)(nil)

func (s *server) CreateProcessDefinition(ctx echo.Context) error {
	if !s.engine.GetPersistence().IsLeader() {
		// if not leader redirect to leader
		return proxyTheRequestToLeader(ctx, s)
	}

	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err.Error())
	}
	process, err := s.engine.LoadFromBytes(data)
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, process)
}

func (s *server) CreateProcessInstance(ctx echo.Context) error {
	if !s.engine.GetPersistence().IsLeader() {
		// if not leader redirect to leader
		return proxyTheRequestToLeader(ctx, s)
	}
	type request struct {
		ProcessDefinitionKey string                 `json:"processDefinitionKey"`
		Variables            map[string]interface{} `json:"variables"`
	}

	var req request
	if err := ctx.Bind(&req); err != nil {
		return returnError(ctx, http.StatusBadRequest, err.Error())
	}
	processDefKey, err := strconv.ParseInt(req.ProcessDefinitionKey, 10, 64)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, fmt.Sprintf("cannot parse process definition key %q: %v", req.ProcessDefinitionKey, err))
	}
	process, err := s.engine.CreateAndRunInstance(processDefKey, req.Variables)
	if err != nil {
		return returnError(ctx, http.StatusInternalServerError, err.Error())
	}

	return s.GetProcessInstance(ctx, process.InstanceKey)
}

func (s *server) GetProcessInstance(ctx echo.Context, processInstanceKey int64) error {
	processInstances := s.engine.GetPersistence().FindProcessInstances(processInstanceKey, -1)
	if len(processInstances) == 0 {
		return returnError(ctx, http.StatusNotFound, fmt.Sprintf("Process instance with key %d not found", processInstanceKey))
	}

	pi := processInstances[0]

	key := fmt.Sprintf("%d", pi.Key)
	processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
	time := time.Unix(0, pi.CreatedAt*int64(time.Second))
	state := api.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
	processInstanceSimple := api.ProcessInstance{
		Key:                  &key,
		ProcessDefinitionKey: &processDefintionKey,
		State:                &state,
		CreatedAt:            &time,
		CaughtEvents:         &pi.CaughtEvents,
		VariableHolder:       &pi.VariableHolder,
		Activities:           &pi.Activities,
	}

	return ctx.JSON(http.StatusOK, processInstanceSimple)

}

func (s *server) GetProcessInstances(ctx echo.Context, params api.GetProcessInstancesParams) error {
	processDefintionKey := int64(-1)
	if params.ProcessDefinitionKey != nil {
		processDefintionKey = int64(*params.ProcessDefinitionKey)
	}
	processInstances := s.engine.GetPersistence().FindProcessInstances(-1, processDefintionKey)

	processInstancesPage := api.ProcessInstancePage{
		Items: &[]api.ProcessInstance{},
	}
	for _, pi := range processInstances {
		key := fmt.Sprintf("%d", pi.Key)
		processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
		time := time.Unix(0, pi.CreatedAt*int64(time.Second))
		state := api.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
		processInstanceSimple := api.ProcessInstance{
			Key:                  &key,
			ProcessDefinitionKey: &processDefintionKey,
			State:                &state,
			CreatedAt:            &time,
			CaughtEvents:         &pi.CaughtEvents,
			VariableHolder:       &pi.VariableHolder,
			Activities:           &pi.Activities,
		}
		*processInstancesPage.Items = append(*processInstancesPage.Items, processInstanceSimple)
	}
	len := len(*processInstancesPage.Items)
	processInstancesPage.Count = &len
	processInstancesPage.Offset = nil
	processInstancesPage.Size = nil

	return ctx.JSON(http.StatusOK, processInstancesPage)
}

func getRedirectLeaderAddress(ctx echo.Context, s *server) (*url.URL, error) {
	leaderUrlString := s.engine.GetPersistence().GetLeaderAddress()
	if leaderUrlString == "" {
		return nil, errors.New("no leader found")
	}
	requestUrl := ctx.Request().URL
	leaderSplit := strings.Split(leaderUrlString, ":")
	leaderPort, err := strconv.Atoi(leaderSplit[1])
	if err != nil {
		return nil, err
	}
	// FIXME: this needs to be independent
	leaderHttpPort := ((s.port / 10) * 10) + leaderPort%10

	redirectUrl := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%d", leaderSplit[0], leaderHttpPort),
		Path:     requestUrl.Path,
		RawQuery: requestUrl.RawQuery,
	}

	return &redirectUrl, nil
}

func proxyTheRequestToLeader(c echo.Context, s *server) error {
	leaderUlr, err := getRedirectLeaderAddress(c, s)
	if err != nil {
		return err
	}
	log.Printf("redirecting request to %s", leaderUlr.String())
	// Create the request to the target server
	req, err := http.NewRequest(c.Request().Method, leaderUlr.String(), c.Request().Body)
	if err != nil {
		return err
	}
	req.Header = c.Request().Header // copy headers

	// Use an HTTP client to send the request to the target server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(name, value)
		}
	}

	// Copy the status code from the target response
	c.Response().WriteHeader(resp.StatusCode)

	// Stream the response body to the client
	_, err = io.Copy(c.Response().Writer, resp.Body)
	return err
}

func (s *server) CompleteJob(ctx echo.Context) error {
	if !s.engine.GetPersistence().IsLeader() {
		// if not leader redirect to leader
		return proxyTheRequestToLeader(ctx, s)
	}
	type completeJobReq struct {
		ProcessInstanceKey int64  `json:"processInstanceKey"`
		JobKey             string `json:"jobKey"`
	}

	var req completeJobReq
	if err := ctx.Bind(&req); err != nil {
		return returnError(ctx, http.StatusBadRequest, err.Error())
	}
	jobKey, err := strconv.ParseInt(req.JobKey, 10, 64)
	if err != nil {
		return returnError(ctx, http.StatusBadRequest, fmt.Sprintf("cannot parse job key %q: %v", req.JobKey, err))
	}
	s.engine.JobCompleteById(jobKey)

	return ctx.NoContent(http.StatusNoContent)
}

func (s *server) GetProcessDefinition(ctx echo.Context, processDefinitionKey int64) error {
	processes := s.engine.GetPersistence().FindProcesses("", processDefinitionKey)

	if len(processes) > 0 {
		version := int(processes[0].Version)
		key := fmt.Sprintf("%d", processes[0].Key)
		ascii85Reader := ascii85.NewDecoder(bytes.NewBuffer([]byte(processes[0].BpmnData)))
		deflateReader := flate.NewReader(ascii85Reader)
		buffer := bytes.Buffer{}
		_, err := io.Copy(&buffer, deflateReader)
		if err != nil {
			return returnError(ctx, http.StatusInternalServerError, err.Error())
		}
		bpmnData := base64.StdEncoding.EncodeToString(buffer.Bytes())
		processDefinitionDetail := &api.ProcessDefinitionDetail{
			Key:           &key,
			Version:       &version,
			BpmnData:      &bpmnData,
			BpmnProcessId: &processes[0].BpmnProcessId,
		}
		return ctx.JSON(http.StatusOK, processDefinitionDetail)
	}
	return returnError(ctx, http.StatusNotFound, "process definition not found")
}

func (s *server) GetProcessDefinitions(ctx echo.Context) error {
	processes := s.engine.GetPersistence().FindProcesses("", -1)
	items := make([]api.ProcessDefinitionSimple, 0)
	result := api.ProcessDefinitionsPage{
		Items: &items,
	}
	for _, p := range processes {
		version := int(p.Version)
		key := fmt.Sprintf("%d", p.Key)
		processDefinitionSimple := api.ProcessDefinitionSimple{
			Key:           &key,
			Version:       &version,
			BpmnProcessId: &p.BpmnProcessId,
		}
		items = append(items, processDefinitionSimple)
	}
	result.Items = &items
	len := len(items)
	result.Count = &len
	result.Offset = nil
	result.Size = nil

	return ctx.JSON(http.StatusOK, result)
}

func (s *server) GetActivities(ctx echo.Context, processInstanceKey int64) error {
	activites := s.engine.GetPersistence().FindActivitiesByProcessInstanceKey(processInstanceKey)
	items := make([]api.Activity, 0)
	result := api.ActivityPage{
		Items: &items,
	}
	for _, a := range activites {
		key := fmt.Sprintf("%d", a.Key)
		time := time.Unix(a.CreatedAt, 0)
		processInstanceKey := fmt.Sprintf("%d", a.ProcessInstanceKey)
		processDefinitionKey := fmt.Sprintf("%d", a.ProcessDefinitionKey)
		activitySimple := api.Activity{
			Key:                  &key,
			ElementId:            &a.ElementId,
			CreatedAt:            &time,
			BpmnElementType:      &a.BpmnElementType,
			ProcessDefinitionKey: &processDefinitionKey,
			ProcessInstanceKey:   &processInstanceKey,
			State:                &a.State,
		}
		items = append(items, activitySimple)
	}
	result.Items = &items
	len := len(items)
	result.Count = &len
	result.Offset = nil
	result.Size = nil

	return ctx.JSON(http.StatusOK, result)
}

func (s *server) GetJobs(ctx echo.Context, processInstanceKey int64) error {
	jobs := s.engine.GetPersistence().FindJobs("", processInstanceKey, int64(-1), []string{})

	items := make([]api.Job, 0)

	result := api.JobPage{
		Items: &items,
	}
	for _, j := range jobs {
		key := fmt.Sprintf("%d", j.Key)
		time := time.Unix(j.CreatedAt, 0)
		processInstanceKey := fmt.Sprintf("%d", j.ProcessInstanceKey)
		elementInstanceKey := fmt.Sprintf("%d", j.ElementInstanceKey)
		//TODO: Needs propper conversion
		state := fmt.Sprintf("%d", j.State)

		jobSimple := api.Job{
			Key:                &key,
			ElementId:          &j.ElementID,
			ElementInstanceKey: &elementInstanceKey,
			CreatedAt:          &time,
			State:              &state,
			ProcessInstanceKey: &processInstanceKey,
		}
		items = append(items, jobSimple)
	}

	result.Items = &items
	len := len(items)
	result.Count = &len
	result.Offset = nil
	result.Size = nil
	return ctx.JSON(http.StatusOK, result)
}

func returnError(ctx echo.Context, code int, message string) error {
	errResponse := message
	return ctx.JSON(code, errResponse)
}
