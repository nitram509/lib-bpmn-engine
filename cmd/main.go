package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api/server"
)

func main() {
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	port := serverFlags.String("port", "8080", "port where to serve traffic")

	portFlagIndex := -1
	for i, arg := range os.Args {
		if arg == "-port" {
			portFlagIndex = i
			break
		}
	}

	if portFlagIndex != -1 {
		if err := serverFlags.Parse(os.Args[portFlagIndex : portFlagIndex+2]); err != nil {
			fmt.Println("Failed to parse server flags:", err)
			return
		}
	}

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

	portInt, err := strconv.Atoi(*port)
	if err != nil {
		panic(err)
	}

	svr := server.NewServer(&engine, portInt)

	api.RegisterHandlers(e, svr)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		if err := e.Server.Shutdown(context.Background()); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	e.Logger.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}
