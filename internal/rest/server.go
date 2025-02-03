package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
)

type Server struct {
	sync.RWMutex
	engine *bpmn_engine.BpmnEngineState
	port   int
	server *http.Server
}

// TODO: do we use non strict interface to implement std lib interface directly and use http.Request to reconstruct grpc calls for proxying?
var _ public.StrictServerInterface = (*Server)(nil)

func NewServer(engine *bpmn_engine.BpmnEngineState, port int) *Server {
	r := chi.NewRouter()
	s := Server{
		engine: engine,
		port:   port,
		server: &http.Server{
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           r,
			Addr:              net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", port)),
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
	return public.CompleteJob201Response{}, nil
}

func (s *Server) GetProcessDefinitions(ctx context.Context, request public.GetProcessDefinitionsRequestObject) (public.GetProcessDefinitionsResponseObject, error) {
	return public.GetProcessDefinitions200JSONResponse{}, nil
}

func (s *Server) GetProcessDefinition(ctx context.Context, request public.GetProcessDefinitionRequestObject) (public.GetProcessDefinitionResponseObject, error) {
	return public.GetProcessDefinition200JSONResponse{}, nil
}

func (s *Server) CreateProcessInstance(ctx context.Context, request public.CreateProcessInstanceRequestObject) (public.CreateProcessInstanceResponseObject, error) {
	return public.CreateProcessInstance200JSONResponse{}, nil
}

func (s *Server) GetProcessInstances(ctx context.Context, request public.GetProcessInstancesRequestObject) (public.GetProcessInstancesResponseObject, error) {
	return public.GetProcessInstances200JSONResponse{}, nil
}

func (s *Server) GetProcessInstance(ctx context.Context, request public.GetProcessInstanceRequestObject) (public.GetProcessInstanceResponseObject, error) {
	return public.GetProcessInstance200JSONResponse{}, nil
}

func (s *Server) GetActivities(ctx context.Context, request public.GetActivitiesRequestObject) (public.GetActivitiesResponseObject, error) {
	return public.GetActivities200JSONResponse{}, nil
}

func (s *Server) GetJobs(ctx context.Context, request public.GetJobsRequestObject) (public.GetJobsResponseObject, error) {
	return public.GetJobs200JSONResponse{}, nil
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
