package rest

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"github.com/pbinitiative/zenbpm/internal/log"
	apierror "github.com/pbinitiative/zenbpm/internal/rest/error"
	"github.com/pbinitiative/zenbpm/internal/rest/middleware"
	"github.com/pbinitiative/zenbpm/internal/rest/public"
	bpmn_engine "github.com/pbinitiative/zenbpm/pkg/bpmn"
	"github.com/pbinitiative/zenbpm/pkg/ptr"
)

type Server struct {
	sync.RWMutex
	engine *bpmn_engine.BpmnEngineState
	addr   string
	server *http.Server
}

// TODO: do we use non strict interface to implement std lib interface directly and use http.Request to reconstruct calls for proxying?
var _ public.StrictServerInterface = (*Server)(nil)

func NewServer(engine *bpmn_engine.BpmnEngineState, addr string) *Server {
	r := chi.NewRouter()
	s := Server{
		engine: engine,
		addr:   addr,
		server: &http.Server{
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           r,
			Addr:              addr,
		},
	}
	r.Use(middleware.Cors())
	r.Route("/v1", func(appContext chi.Router) {
		h := public.HandlerFromMux(public.NewStrictHandlerWithOptions(&s, []nethttp.StrictHTTPMiddlewareFunc{}, public.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeError(w, r, http.StatusInternalServerError, apierror.ApiError{
					Message: err.Error(),
					Type:    "ERROR",
				})
			},
		}), appContext)
		appContext.Mount("/", h)
	})
	return &s
}

func (s *Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Error starting server: %s", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Error("Error stopping server: %s", err)
	}
}

func (s *Server) CreateProcessDefinition(ctx context.Context, request public.CreateProcessDefinitionRequestObject) (public.CreateProcessDefinitionResponseObject, error) {
	if !s.engine.GetPersistence().IsLeader() {
		// if not leader redirect to leader
		// proxyTheRequestToLeader(ctx, s)
		return nil, fmt.Errorf("not leader")
	}

	data, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	process, err := s.engine.LoadFromBytes(data)
	if err != nil {
		return nil, err
	}
	return public.CreateProcessDefinition200JSONResponse{
		ProcessDefinitionKey: &process.BpmnProcessId,
	}, nil
}

func (s *Server) CompleteJob(ctx context.Context, request public.CompleteJobRequestObject) (public.CompleteJobResponseObject, error) {
	s.engine.JobCompleteById(request.Body.JobKey)
	return public.CompleteJob201Response{}, nil
}

func (s *Server) GetProcessDefinitions(ctx context.Context, request public.GetProcessDefinitionsRequestObject) (public.GetProcessDefinitionsResponseObject, error) {
	processes := s.engine.GetPersistence().FindProcesses("", -1)
	items := make([]public.ProcessDefinitionSimple, 0)
	result := public.ProcessDefinitionsPage{
		Items: &items,
	}
	for _, p := range processes {
		version := int(p.Version)
		key := fmt.Sprintf("%d", p.Key)
		processDefinitionSimple := public.ProcessDefinitionSimple{
			Key:           &key,
			Version:       &version,
			BpmnProcessId: &p.BpmnProcessId,
		}
		items = append(items, processDefinitionSimple)
	}
	result.Items = &items
	total := len(items)
	result.Count = &total
	result.Offset = nil
	result.Size = nil

	return public.GetProcessDefinitions200JSONResponse(result), nil
}

func (s *Server) GetProcessDefinition(ctx context.Context, request public.GetProcessDefinitionRequestObject) (public.GetProcessDefinitionResponseObject, error) {
	processes := s.engine.GetPersistence().FindProcesses("", request.ProcessDefinitionKey)
	if len(processes) == 0 {
		return public.GetProcessDefinition200JSONResponse{}, nil
	}

	version := int(processes[0].Version)
	key := fmt.Sprintf("%d", processes[0].Key)
	ascii85Reader := ascii85.NewDecoder(bytes.NewBuffer([]byte(processes[0].BpmnData)))
	deflateReader := flate.NewReader(ascii85Reader)
	buffer := bytes.Buffer{}
	_, err := io.Copy(&buffer, deflateReader)
	if err != nil {
		return nil, err
	}
	bpmnData := base64.StdEncoding.EncodeToString(buffer.Bytes())
	processDefinitionDetail := public.ProcessDefinitionDetail{
		ProcessDefinitionSimple: public.ProcessDefinitionSimple{
			BpmnProcessId: &processes[0].BpmnProcessId,
			Key:           &key,
			Version:       &version,
		},
		BpmnData: &bpmnData,
	}
	return public.GetProcessDefinition200JSONResponse(processDefinitionDetail), nil
}

func (s *Server) CreateProcessInstance(ctx context.Context, request public.CreateProcessInstanceRequestObject) (public.CreateProcessInstanceResponseObject, error) {
	variables := make(map[string]interface{})
	if request.Body.Variables != nil {
		variables = *request.Body.Variables
	}
	process, err := s.engine.CreateAndRunInstance(request.Body.ProcessDefinitionKey, variables)
	if err != nil {
		return nil, err
	}
	instanceDetail, err := s.getProcessInstance(ctx, process.InstanceKey)
	if err != nil {
		return nil, err
	}
	return public.CreateProcessInstance200JSONResponse(*instanceDetail), nil
}

func (s *Server) GetProcessInstances(ctx context.Context, request public.GetProcessInstancesRequestObject) (public.GetProcessInstancesResponseObject, error) {
	processDefintionKey := int64(-1)
	if request.Params.ProcessDefinitionKey != nil {
		processDefintionKey = *request.Params.ProcessDefinitionKey
	}
	processInstances := s.engine.GetPersistence().FindProcessInstances(-1, processDefintionKey)
	processInstancesPage := public.ProcessInstancePage{
		Items: &[]public.ProcessInstance{},
	}
	for _, pi := range processInstances {
		processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
		time := time.Unix(0, pi.CreatedAt*int64(time.Second))
		state := public.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
		processInstanceSimple := public.ProcessInstance{
			Key:                  pi.Key,
			ProcessDefinitionKey: processDefintionKey,
			State:                state,
			CreatedAt:            &time,
			CaughtEvents:         &pi.CaughtEvents,
			VariableHolder:       &pi.VariableHolder,
			Activities:           &pi.Activities,
		}
		*processInstancesPage.Items = append(*processInstancesPage.Items, processInstanceSimple)
	}
	total := len(*processInstancesPage.Items)
	processInstancesPage.Count = &total
	processInstancesPage.Offset = nil
	processInstancesPage.Size = nil
	return public.GetProcessInstances200JSONResponse{
		ProcessInstances: &[]public.ProcessInstancePage{processInstancesPage},
		Total:            total,
	}, nil
}

func (s *Server) getProcessInstance(ctx context.Context, key int64) (*public.ProcessInstance, error) {
	processInstances := s.engine.GetPersistence().FindProcessInstances(key, -1)
	if len(processInstances) == 0 {
		return nil, fmt.Errorf("process instance with key %d not found", key)
	}
	pi := processInstances[0]
	processDefintionKey := fmt.Sprintf("%d", pi.ProcessDefinitionKey)
	time := time.Unix(0, pi.CreatedAt*int64(time.Second))
	state := public.ProcessInstanceState(fmt.Sprintf("%d", pi.State))
	processInstanceSimple := public.ProcessInstance{
		Key:                  pi.Key,
		ProcessDefinitionKey: processDefintionKey,
		State:                state,
		CreatedAt:            &time,
		CaughtEvents:         &pi.CaughtEvents,
		VariableHolder:       &pi.VariableHolder,
		Activities:           &pi.Activities,
	}
	return &processInstanceSimple, nil
}

func (s *Server) GetProcessInstance(ctx context.Context, request public.GetProcessInstanceRequestObject) (public.GetProcessInstanceResponseObject, error) {
	processInstances := s.engine.GetPersistence().FindProcessInstances(request.ProcessInstanceKey, -1)
	if len(processInstances) == 0 {
		return nil, fmt.Errorf("process instance with key %d not found", request.ProcessInstanceKey)
	}
	pi := processInstances[0]

	processInstanceSimple := public.ProcessInstance{
		Key:                  pi.Key,
		ProcessDefinitionKey: fmt.Sprintf("%d", pi.ProcessDefinitionKey),
		State:                public.ProcessInstanceState(fmt.Sprintf("%d", pi.State)),
		CreatedAt:            ptr.To(time.Unix(0, pi.CreatedAt*int64(time.Second))),
		CaughtEvents:         &pi.CaughtEvents,
		VariableHolder:       &pi.VariableHolder,
		Activities:           &pi.Activities,
	}
	return public.GetProcessInstance200JSONResponse(processInstanceSimple), nil
}

func (s *Server) GetActivities(ctx context.Context, request public.GetActivitiesRequestObject) (public.GetActivitiesResponseObject, error) {
	activites := s.engine.GetPersistence().FindActivitiesByProcessInstanceKey(request.ProcessInstanceKey)
	items := make([]public.Activity, 0)
	result := public.ActivityPage{
		Items: &items,
	}
	for _, a := range activites {
		key := fmt.Sprintf("%d", a.Key)
		time := time.Unix(a.CreatedAt, 0)
		processInstanceKey := fmt.Sprintf("%d", a.ProcessInstanceKey)
		processDefinitionKey := fmt.Sprintf("%d", a.ProcessDefinitionKey)
		activitySimple := public.Activity{
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
	return public.GetActivities200JSONResponse(result), nil
}

func (s *Server) GetJobs(ctx context.Context, request public.GetJobsRequestObject) (public.GetJobsResponseObject, error) {
	jobs := s.engine.GetPersistence().FindJobs("", request.ProcessInstanceKey, int64(-1), []string{})
	items := make([]public.Job, 0)
	result := public.JobPage{
		Items: &items,
	}
	for _, j := range jobs {
		key := fmt.Sprintf("%d", j.Key)
		time := time.Unix(j.CreatedAt, 0)
		processInstanceKey := fmt.Sprintf("%d", j.ProcessInstanceKey)
		elementInstanceKey := fmt.Sprintf("%d", j.ElementInstanceKey)
		//TODO: Needs propper conversion
		state := fmt.Sprintf("%d", j.State)
		jobSimple := public.Job{
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
	return public.GetJobs200JSONResponse(result), nil
}

func writeError(w http.ResponseWriter, r *http.Request, status int, resp interface{}) {
	w.WriteHeader(status)
	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("Server error: %s", err)
	} else {
		w.Write(body)
	}
}
