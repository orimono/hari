package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/orimono/ito"
	"github.com/shirou/gopsutil/v4/disk"
)

type DiskCollector struct{}

func (c *DiskCollector) Name() string {
	return "disk"
}

func (c *DiskCollector) Capability() ito.Capability {
	return ito.Capability{
		Kind:      "collector",
		Version:   "0.1.0",
		Platforms: []string{"linux", "darwin", "windows"},
	}
}

func (c *DiskCollector) Interval() time.Duration {
	return time.Second * 30
}

func (c *DiskCollector) Collect(ctx context.Context) (any, error) {
	parts, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	partitions := make([]ito.DiskPartition, 0, len(parts))
	for _, p := range parts {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			slog.Warn("failed to get disk usage", "mountpoint", p.Mountpoint, "err", err)
			continue
		}
		partitions = append(partitions, ito.DiskPartition{
			Mountpoint:  p.Mountpoint,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	slog.Debug("collected disk", "partitions", len(partitions))
	return ito.DiskMetrics{Partitions: partitions}, nil
}
