package main

import (
	"flag"
	"net"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api/server"
)

func main() {
	port := flag.String("port", "8080", "port where to serve traffic")

	e := echo.New()

	engine := bpmn_engine.New()
	// TODO rework handlers
	emptyHandler := func(job bpmn_engine.ActivatedJob) {
	}
	engine.NewTaskHandler().Type("foo").Handler(emptyHandler)
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	svr := server.NewServer(&engine)

	api.RegisterHandlers(e, svr)
	e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}
