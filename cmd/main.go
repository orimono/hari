package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/orimono/shutter/internal/collector"
	"github.com/orimono/shutter/internal/collector/subsystem"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/logger"
	"github.com/orimono/shutter/internal/reporter"
	"github.com/orimono/shutter/internal/ws"
	"github.com/orimono/ito"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel)
	client := ws.NewClient(cfg)
	go client.Serve(ctx)

	reporter := reporter.NewReporter(client)

	go func() {
		data, err := os.ReadFile("./testdata.json")

		if err != nil {
			slog.Error("Failed to read JSON from file")
			return
		}

		var joinPacket = &ito.JoinPacket{}
		json.Unmarshal(data, joinPacket)

		joinPacket.NodeID, err = ito.GenerateNodeID(joinPacket.NodeID)
		if err != nil {
			slog.Error("Failed to read JSON from file")
			return
		}

		modifiedData, marshalErr := json.Marshal(joinPacket)
		if marshalErr != nil {
			slog.Error("Failed to marshal")
			return
		}
		reporter.Send(modifiedData)
	}()

	manager := collector.NewManager()
	manager.AddCollector(&subsystem.MemoryCollector{})
	go manager.Start(ctx)

	select {}
}
