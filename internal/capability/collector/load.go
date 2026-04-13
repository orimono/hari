package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/orimono/ito"
	"github.com/shirou/gopsutil/v4/load"
)

type LoadCollector struct{}

func (c *LoadCollector) Name() string {
	return "load"
}

func (c *LoadCollector) Capability() ito.Capability {
	return ito.Capability{
		Kind:      "collector",
		Version:   "0.1.0",
		Platforms: []string{"linux", "darwin"},
	}
}

func (c *LoadCollector) Interval() time.Duration {
	return time.Second * 15
}

func (c *LoadCollector) Collect(ctx context.Context) (any, error) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	slog.Debug("collected load", "load1", avg.Load1, "load5", avg.Load5, "load15", avg.Load15)
	return ito.LoadMetrics{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}
