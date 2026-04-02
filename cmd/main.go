package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"runtime"

	"github.com/orimono/ito"
	"github.com/orimono/shutter/internal/collector"
	"github.com/orimono/shutter/internal/collector/subsystem"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/logger"
	"github.com/orimono/shutter/internal/reporter"
	"github.com/orimono/shutter/internal/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel)

	manager := collector.NewManager()
	manager.AddCollector(&subsystem.MemoryCollector{})

	client := ws.NewClient(cfg)
	go client.Run(ctx)

	rep := reporter.NewReporter(client)

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			slog.Error("failed to get hostname", "err", err)
			return
		}

		joinPacket := &ito.JoinPacket{
			Hostname:     hostname,
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Tags:         cfg.Tags,
			TaskManifest: manager.Manifest(),
		}

		joinPacket.NodeID, err = ito.GenerateNodeID("")
		if err != nil {
			slog.Error("failed to generate node ID", "err", err)
			return
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

	<-ctx.Done()
}
