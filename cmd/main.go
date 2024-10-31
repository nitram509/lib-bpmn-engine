package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/api/server"
	rqlitePersitence "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/persistence/rqlite"
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

	cfg := rqlitePersitence.ParseConfig()

	partitions := cfg.Partitions
	engines := make([]*bpmn_engine.BpmnEngineState, partitions)
	nodeId, err := strconv.Atoi(cfg.NodeID)
	if err != nil {
		panic(err)
	}

	for i := 0; i < partitions; i++ {
		cfgCopy := *cfg
		cfgCopy.RaftAddr = getPartitionPort(cfgCopy.RaftAddr, i)
		cfgCopy.RaftAdv = getPartitionPort(cfgCopy.RaftAdv, i)
		cfgCopy.HTTPAddr = getPartitionPort(cfgCopy.HTTPAddr, i)
		cfgCopy.HTTPAdv = getPartitionPort(cfgCopy.HTTPAdv, i)
		cfgCopy.DataPath = fmt.Sprintf("%s/%d", cfgCopy.DataPath, i)

		joinAddrs := strings.Split(cfgCopy.JoinAddrs, ",")

		for j := 0; j < len(joinAddrs); j++ {
			joinAddrs[j] = getPartitionPort(joinAddrs[j], i)
		}
		cfgCopy.JoinAddrs = strings.Join(joinAddrs, ",")

		cfgCopy.NodeID = fmt.Sprintf("%d", nodeId*100+i)
		log.Printf("Node id: %s", cfgCopy.NodeID)
		engine := bpmn_engine.NewWithConfig(&cfgCopy)
		// TODO rework handlers
		emptyHandler := func(job bpmn_engine.ActivatedJob) {
		}
		engine.NewTaskHandler().Type("foo").Handler(emptyHandler)
		engines[i] = &engine
	}

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

	svr := server.NewServer(engines, portInt)

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

func getPartitionPort(address string, partition int) string {
	hp := strings.Split(address, ":")
	if len(hp) < 2 {
		return address
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		panic(err)
	}
	port += partition
	hp[len(hp)-1] = strconv.Itoa(port)
	return strings.Join(hp, ":")
}
