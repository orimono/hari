package collector

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

func (c *MemoryCollector) Capability() ito.Capability {
	return ito.Capability{
		Kind:      "collector",
		Version:   "0.1.0",
		Platforms: []string{"linux", "darwin", "windows"},
	}
}

func (c *MemoryCollector) Interval() time.Duration {
	return time.Second * 5
}

func (c *MemoryCollector) Collect(ctx context.Context) (any, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	slog.Debug("collected memory", "used_percent", v.UsedPercent)
	return ito.MemoryMetrics{
		Total:       v.Total,
		Used:        v.Used,
		Free:        v.Free,
		UsedPercent: v.UsedPercent,
	}, nil
}
