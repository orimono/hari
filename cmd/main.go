package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"runtime"

	"errors"

	"github.com/orimono/ito"

	"github.com/orimono/shutter/internal/capability"
	"github.com/orimono/shutter/internal/capability/collector"
	"github.com/orimono/shutter/internal/capability/executor"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/dispatcher"
	"github.com/orimono/shutter/internal/logger"
	"github.com/orimono/shutter/internal/reporter"
	"github.com/orimono/shutter/internal/store"
	"github.com/orimono/shutter/internal/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel)

	nodeID, err := ito.GenerateNodeID("")
	if err != nil {
		slog.Error("failed to generate node ID", "err", err)
		return
	}

	storePath := cfg.StorePath
	if storePath == "" {
		storePath = "shutter.db"
	}
	execStore, err := store.NewExecutorStore(storePath)
	if err != nil {
		slog.Error("failed to open executor store", "err", err)
		return
	}

	manager := capability.NewManager(nodeID)
	manager.AddCollector(&collector.MemoryCollector{})
	manager.AddCollector(&collector.CPUCollector{})
	manager.AddCollector(&collector.DiskCollector{})
	manager.AddCollector(&collector.NetworkCollector{})
	manager.AddCollector(&collector.LoadCollector{})
	manager.AddExecutor(&executor.ServiceRestartExecutor{})
	manager.AddExecutor(&executor.ShutdownExecutor{})
	manager.AddExecutor(&executor.RebootExecutor{})

	// load dynamic executors persisted from previous sessions
	dynamic, err := execStore.LoadAll()
	if err != nil {
		slog.Warn("failed to load dynamic executors", "err", err)
	}
	for _, e := range dynamic {
		manager.AddExecutor(e)
		slog.Info("loaded dynamic executor", "kind", e.Name())
	}

	d := dispatcher.New(manager, execStore)
	client := ws.NewClient(cfg, d)
	go client.Run(ctx)

	rep := reporter.NewReporter(client)

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			slog.Error("failed to get hostname", "err", err)
			return
		}

		joinPacket := &ito.JoinPacket{
			NodeID:       nodeID,
			Hostname:     hostname,
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Tags:         cfg.Tags,
			TaskManifest: manager.Manifest(),
		}

		data, err := json.Marshal(joinPacket)
		if err != nil {
			slog.Error("failed to marshal JoinPacket", "err", err)
			return
		}

		<-client.Ready()
		rep.Send(data)
	}()

	go manager.Start(ctx)

	go func() {
		for t := range manager.Out() {
			data, err := json.Marshal(t)
			if err != nil {
				slog.Error("failed to marshal telemetry", "err", err)
				continue
			}
			if err := client.Send(data); err != nil {
				if errors.Is(err, ws.ErrNoSession) {
					slog.Debug("failed to send telemetry", "err", err)
				} else {
					slog.Warn("failed to send telemetry", "err", err)
				}
			}
		}
	}()

	<-ctx.Done()
}
