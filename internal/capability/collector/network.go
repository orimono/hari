package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/orimono/ito"
	"github.com/shirou/gopsutil/v4/net"
)

type NetworkCollector struct{}

func (c *NetworkCollector) Name() string {
	return "net"
}

func (c *NetworkCollector) Capability() ito.Capability {
	return ito.Capability{
		Kind:      "collector",
		Version:   "0.1.0",
		Platforms: []string{"linux", "darwin", "windows"},
	}
}

func (c *NetworkCollector) Interval() time.Duration {
	return time.Second * 10
}

func (c *NetworkCollector) Collect(ctx context.Context) (any, error) {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	ifaces := make([]ito.NetworkInterface, 0, len(counters))
	for _, c := range counters {
		ifaces = append(ifaces, ito.NetworkInterface{
			Name:        c.Name,
			BytesSent:   c.BytesSent,
			BytesRecv:   c.BytesRecv,
			PacketsSent: c.PacketsSent,
			PacketsRecv: c.PacketsRecv,
		})
	}

	slog.Debug("collected network", "interfaces", len(ifaces))
	return ito.NetworkMetrics{Interfaces: ifaces}, nil
}
