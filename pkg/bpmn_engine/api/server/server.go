package server

import (
	"bytes"
	"compress/flate"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/json"
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

	"github.com/bwmarrin/snowflake"
	"github.com/labstack/echo/v4"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api"
	"golang.org/x/exp/rand"
)

type server struct {
	sync.RWMutex
	engines []*bpmn_engine.BpmnEngineState
	port    int
}

func NewServer(engines []*bpmn_engine.BpmnEngineState, port int) *server {
	return &server{
		engines: engines,
		port:    port,
	}
}

// Ensure that we implement the server interface
var _ api.ServerInterface = (*server)(nil)

func (s *server) CreateProcessDefinition(ctx echo.Context) error {
	// This needs to be broadcasted to all engines
	log.Printf("===========================================\nSTART Creating process definition")

	var process *bpmn_engine.ProcessInfo
	partition := ctx.Request().Header.Get("X-Partition")
	processKey := ctx.Request().Header.Get("X-ProcessKey")

	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	// Handle initial creation in partition one
	// Request from the client
	if partition == "" && processKey == "" {
		// Check i f I am leader of partition 1
		if !s.engines[0].GetPersistence().IsLeader() {
			log.Printf("INITIAL: REDIRECTING: creating process definition in partition %d", 0)
			proxyTheRequestToLeader(ctx, s, 0, data)
		} else {
			log.Printf("INITIAL: HANDING: creating process definition in partition %d", 0)
			process, err = handleProcessDefinitionCreation(ctx, data, s.engines[0], nil)
			if err != nil {
				return returnError(ctx, http.StatusInternalServerError, err.Error())
			}
		}
		// Request redirected from node that is not partition 0 leader
	} else if partition == "0" && processKey == "" {
		log.Printf("REDIRECTED: INITIAL: HANDING: creating process definition in partition %d", 0)
		process, err = handleProcessDefinitionCreation(ctx, data, s.engines[0], nil)
		if err != nil {
			return returnError(ctx, http.StatusInternalServerError, err.Error())
		}
		// Request redirected from node that is partition 0 leader for handing other partitions
	} else if partition != "" && processKey != "" {
		log.Printf("REDIRECTED: HANDING: creating process definition in partition %d", 0)
		process, err = handleProcessDefinitionCreation(ctx, data, s.engines[0], nil)
		if err != nil {
			return returnError(ctx, http.StatusInternalServerError, err.Error())
		}
	}

	// Handle propagation to other partitions
	if process != nil {
		for i := 1; i < len(s.engines); i++ {
			eng := s.engines[i]
			if !eng.GetPersistence().IsLeader() {
				// if not leader redirect to leader and pass the processKey
				ctx.Request().Header.Add("X-ProcessKey", fmt.Sprintf("%d", process.ProcessKey))
				proxyTheRequestToLeader(ctx, s, i, data)

			} else {
				log.Printf("creating process definition in partition %d", i)
				process, err = handleProcessDefinitionCreation(ctx, data, eng, &(process.ProcessKey))

				if err != nil {
					return returnError(ctx, http.StatusInternalServerError, err.Error())
				}
			}
		}
	}

	log.Printf("===========================================\nDONE Creating process definition ")
	return ctx.JSON(http.StatusOK, process)
}

func handleProcessDefinitionCreation(ctx echo.Context, data []byte, engine *bpmn_engine.BpmnEngineState, processKey *int64) (*bpmn_engine.ProcessInfo, error) {
	var process *bpmn_engine.ProcessInfo
	var err error
	if processKey != nil {
		process, err = engine.LoadFromBytesWithKey(data, *processKey)
	} else {
		process, err = engine.LoadFromBytes(data)
	}
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (s *server) CreateProcessInstance(ctx echo.Context) error {
	// Create the process instance in partition where I am a leader. Is leader is not true, redirect random paritions leader
	log.Printf("===========================================\nSTART Creating process instance")
	for i, eng := range s.engines {
		if eng.GetPersistence().IsLeader() {
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
			log.Printf("partition %d addr %s eng %s", i, eng.GetPersistence().GetLeaderAddress(), eng.Name())
			process, err := eng.CreateAndRunInstance(processDefKey, req.Variables)
			if err != nil {
				return returnError(ctx, http.StatusInternalServerError, err.Error())
			}

			return s.GetProcessInstance(ctx, process.InstanceKey)
		}
	}

	// Redirect to a random partition leader if no leader is found in current instance
	randIndex := getRandomPartitionIndex(s)
	return proxyTheRequestToLeader(ctx, s, randIndex)
}

func getRandomPartitionIndex(s *server) int {
	rand.Seed(uint64(time.Now().UnixNano()))
	randIndex := rand.Intn(len(s.engines))
	return randIndex
}

func isPartitionsLeader(s *server, partition int) bool {
	return s.engines[partition].GetPersistence().IsLeader()
}

func (s *server) GetProcessInstance(ctx echo.Context, processInstanceKey int64) error {
	parsedNodeObject := bpmn_engine.ParseSnowflake(snowflake.ID(processInstanceKey))
	log.Printf("Parsed snowflake ID: %+v", parsedNodeObject)
	if !s.engines[parsedNodeObject.Partition].GetPersistence().IsLeader() {
		log.Printf("REDIRECTED: HANDING: getting process instance in partition %d", parsedNodeObject.Partition)
		return proxyTheRequestToLeader(ctx, s, int(parsedNodeObject.Partition))
	}

	log.Printf("HANDING: Getting process instance in partition %d", parsedNodeObject.Partition)

	processInstances := s.engines[parsedNodeObject.Partition].GetPersistence().FindProcessInstances(processInstanceKey, -1)
	if len(processInstances) == 0 {
		return returnError(ctx, http.StatusNotFound, fmt.Sprintf("Process instance with key %d not found", processInstanceKey))
	}

	pi := processInstances[0]

	key := fmt.Sprintf("%d", pi.Key)
	processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
	createdAt := time.Unix(0, pi.CreatedAt*int64(time.Second))
	completedAt := time.Unix(0, pi.CompletedAt*int64(time.Second))
	state := api.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
	processInstanceSimple := api.ProcessInstance{
		Key:                  &key,
		ProcessDefinitionKey: &processDefintionKey,
		State:                &state,
		CreatedAt:            &createdAt,
		CompletedAt:          &completedAt,
		CaughtEvents:         &pi.CaughtEvents,
		VariableHolder:       &pi.VariableHolder,
		Activities:           &pi.Activities,
	}

	ctx.Response().Header().Set("X-Partition", fmt.Sprintf("%d", parsedNodeObject.Partition))
	return ctx.JSON(http.StatusOK, processInstanceSimple)

}

func (s *server) GetProcessInstances(ctx echo.Context, params api.GetProcessInstancesParams) error {
	processDefintionKey := int64(-1)
	if params.ProcessDefinitionKey != nil {
		processDefintionKey = int64(*params.ProcessDefinitionKey)
	}
	processInstancesPage := api.ProcessInstancePage{
		Items: &[]api.ProcessInstance{},
	}
	// Reads the instances from all the partitions
	// TODO: The paging will be hard
	for _, engine := range s.engines {
		partitionProcessInstances := engine.GetPersistence().FindProcessInstances(-1, processDefintionKey)
		for _, pi := range partitionProcessInstances {
			key := fmt.Sprintf("%d", pi.Key)
			processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
			createdAt := time.Unix(0, pi.CreatedAt*int64(time.Second))
			completedAt := time.Unix(0, pi.CompletedAt*int64(time.Second))
			state := api.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
			processInstanceSimple := api.ProcessInstance{
				Key:                  &key,
				ProcessDefinitionKey: &processDefintionKey,
				State:                &state,
				CreatedAt:            &createdAt,
				CompletedAt:          &completedAt,
				CaughtEvents:         &pi.CaughtEvents,
				VariableHolder:       &pi.VariableHolder,
				Activities:           &pi.Activities,
			}
			*processInstancesPage.Items = append(*processInstancesPage.Items, processInstanceSimple)
		}
	}
	len := len(*processInstancesPage.Items)
	processInstancesPage.Count = &len
	processInstancesPage.Offset = nil
	processInstancesPage.Size = nil

	return ctx.JSON(http.StatusOK, processInstancesPage)
}

func getRedirectLeaderAddress(ctx echo.Context, s *server, partition int) (*url.URL, error) {
	leaderUrlString := s.engines[partition].GetPersistence().GetLeaderAddress()
	if leaderUrlString == "" {
		return nil, errors.New("no leader found")
	}
	requestUrl := ctx.Request().URL
	leaderSplit := strings.Split(leaderUrlString, ":")
	// leaderPort, err := strconv.Atoi(leaderSplit[1])
	// if err != nil {
	// return nil, err
	// }
	// FIXME: this needs to be independent
	// leaderHttpPort := ((s.port / 10) * 10) + leaderPort%10
	leaderHttpPort := s.port

	redirectUrl := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%d", leaderSplit[0], leaderHttpPort),
		Path:     requestUrl.Path,
		RawQuery: requestUrl.RawQuery,
	}

	return &redirectUrl, nil
}

func proxyTheRequestToLeader(c echo.Context, s *server, partition int, body ...[]byte) error {
	leaderUlr, err := getRedirectLeaderAddress(c, s, partition)
	if err != nil {
		return err
	}
	log.Printf("redirecting request to %s", leaderUlr.String())
	// Create the request to the target server
	var req *http.Request
	if len(body) == 0 {
		req, err = http.NewRequest(c.Request().Method, leaderUlr.String(), c.Request().Body)
	} else {
		req, err = http.NewRequest(c.Request().Method, leaderUlr.String(), io.NopCloser(bytes.NewBuffer(body[0])))

	}
	if err != nil {
		return err
	}
	req.Header = c.Request().Header // copy headers
	req.Header.Set("X-Partition", fmt.Sprintf("%d", partition))

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

	parsedNodeObject := bpmn_engine.ParseSnowflake(snowflake.ID(jobKey))
	if !isPartitionsLeader(s, int(parsedNodeObject.Partition)) {
		json, _ := json.Marshal(req)
		return proxyTheRequestToLeader(ctx, s, int(parsedNodeObject.Partition), json)
	}

	s.engines[int(parsedNodeObject.Partition)].JobCompleteById(jobKey)

	return ctx.NoContent(http.StatusNoContent)
}

func (s *server) GetProcessDefinition(ctx echo.Context, processDefinitionKey int64) error {
	processes := s.engines[getRandomPartitionIndex(s)].GetPersistence().FindProcesses("", processDefinitionKey)

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
	processes := s.engines[getRandomPartitionIndex(s)].GetPersistence().FindProcesses("", -1)
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
	parsedNodeObject := bpmn_engine.ParseSnowflake(snowflake.ID(processInstanceKey))
	activites := s.engines[parsedNodeObject.Partition].GetPersistence().FindActivitiesByProcessInstanceKey(processInstanceKey)
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
	parsedNodeObject := bpmn_engine.ParseSnowflake(snowflake.ID(processInstanceKey))
	jobs := s.engines[parsedNodeObject.Partition].GetPersistence().FindJobs("", processInstanceKey, int64(-1), []string{})

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

func (s *server) GetClusterInfo(ctx echo.Context) error {
	partitions := make([]api.ClusterPartition, len(s.engines))
	for i, eng := range s.engines {
		leader := eng.GetPersistence().GetLeaderAddress()
		members := strings.Split(eng.GetPersistence().GetJoinAddresses(), ",")
		id := fmt.Sprintf("%d", i)
		clusterPartition := api.ClusterPartition{
			Id:      &id,
			Leader:  &leader,
			Members: &members,
		}
		partitions[i] = clusterPartition
	}

	clusterInfo := api.ClusterInfo{
		Partitions: &partitions,
	}

	return ctx.JSON(http.StatusOK, clusterInfo)
}

func (s *server) Rebalance(ctx echo.Context) error {
	type LeadershipSummary map[string][]int

	leadershipSummary := LeadershipSummary{}

	for i, eng := range s.engines {
		leader := strings.Split(eng.GetPersistence().GetLeaderAddress(), ":")[0]
		leadershipSummary[leader] = append(leadershipSummary[leader], i)
		if len(leadershipSummary[leader]) > 1 {
			if eng.GetPersistence().IsLeader() {
				eng.GetPersistence().StepdownAsLeader()
			} else {
				proxyTheRequestToLeader(ctx, s, i)
			}
			break
		}
	}
	log.Printf("Leadership summary: %v", leadershipSummary)
	return ctx.NoContent(http.StatusNoContent)

}

func returnError(ctx echo.Context, code int, message string) error {
	errResponse := message
	return ctx.JSON(code, errResponse)
}
