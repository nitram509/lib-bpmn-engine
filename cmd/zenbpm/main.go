package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pbinitiative/zenbpm/internal/cluster"
	"github.com/pbinitiative/zenbpm/internal/config"
	"github.com/pbinitiative/zenbpm/internal/log"
	"github.com/pbinitiative/zenbpm/internal/profile"
	"github.com/pbinitiative/zenbpm/internal/rest"
	bpmn_engine "github.com/pbinitiative/zenbpm/pkg/bpmn"
)

func main() {
	profile.InitProfile()
	log.Init()

	appContext, ctxCancel := context.WithCancel(context.Background())

	conf := config.InitConfig()

	zenNode, err := cluster.StartZenNode(appContext, conf)
	if err != nil {
		log.Error("Failed to start Zen node: %s", err)
		os.Exit(1)
	}

	engine := bpmn_engine.New(zenNode)
	// TODO rework handlers
	emptyHandler := func(job bpmn_engine.ActivatedJob) {
	}
	engine.NewTaskHandler().Type("foo").Handler(emptyHandler)

	svr := rest.NewServer(&engine, conf.Server.Addr)
	svr.Start()

	appStop := make(chan os.Signal, 2)
	signal.Notify(appStop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	handleSigterm(appStop, appContext)

	ctxCancel()
	// cleanup
	svr.Stop(appContext)
	err = zenNode.Stop()
	if err != nil {
		log.Error("failed to properly stop zen node: %s", err)
	}
}

func handleSigterm(appStop chan os.Signal, ctx context.Context) {
	signal.Notify(appStop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-appStop
	log.Infof(ctx, "Received %s. Shutting down", sig.String())
}
