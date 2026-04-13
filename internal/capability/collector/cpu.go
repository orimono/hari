package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/orimono/ito"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUCollector struct{}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func (c *CPUCollector) Capability() ito.Capability {
	return ito.Capability{
		Kind:      "collector",
		Version:   "0.1.0",
		Platforms: []string{"linux", "darwin", "windows"},
	}
}

func (c *CPUCollector) Interval() time.Duration {
	return time.Second * 5
}

func (c *CPUCollector) Collect(ctx context.Context) (any, error) {
	total, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, err
	}

	perCPU, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, err
	}

	usedPercent := 0.0
	if len(total) > 0 {
		usedPercent = total[0]
	}

	slog.Debug("collected cpu", "used_percent", usedPercent)
	return ito.CPUMetrics{
		UsedPercent: usedPercent,
		PerCPU:      perCPU,
	}, nil
}
