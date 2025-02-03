package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pbinitiative/zenbpm/internal/log"
	"github.com/pbinitiative/zenbpm/internal/rest"
	bpmn_engine "github.com/pbinitiative/zenbpm/pkg/bpmn"
)

func main() {
	appContext, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

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

	engine := bpmn_engine.New()
	// TODO rework handlers
	emptyHandler := func(job bpmn_engine.ActivatedJob) {
	}
	engine.NewTaskHandler().Type("foo").Handler(emptyHandler)

	portInt, err := strconv.Atoi(*port)
	if err != nil {
		panic(err)
	}

	svr := rest.NewServer(&engine, portInt)

	svr.Start()
	appStop := make(chan os.Signal, 2)
	signal.Notify(appStop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	handleSigterm(appStop, appContext)
	svr.Stop(appContext)
}

func handleSigterm(appStop chan os.Signal, ctx context.Context) {
	signal.Notify(appStop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-appStop
	log.Infof(ctx, "Received %s. Shutting down", sig.String())
}
