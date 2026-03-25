package subsystem

import (
	"context"
	"log/slog"
	"time"

	"github.com/orimono/ito"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemoryCollector struct{}

func (c *MemoryCollector) Name() string {
	return "mem"
}

func (c *MemoryCollector) Interval() time.Duration {
	return time.Second * 2
}

func (c *MemoryCollector) Collect(ctx context.Context) (any, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to get virtual memory data", "error", err)
	}

	slog.Info("Collected memory data.", "data", v)
	return ito.MemoryMetrics{}, nil
}
