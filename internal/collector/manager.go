package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/orimono/ito"
)

type Manager struct {
	nodeID     string
	collectors []Collector
	out        chan<- ito.Telemetry
}

func NewManager() *Manager {
	nodeID, err := ito.GenerateNodeID("app")
	if err != nil {
		panic("Failed to get nodeID")
	}
	return &Manager{
		nodeID:     nodeID,
		collectors: make([]Collector, 0),
		out:        make(chan ito.Telemetry, 100),
	}
}

func (m *Manager) AddCollector(c Collector) {
	m.collectors = append(m.collectors, c)
}

func (m *Manager) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for _, c := range m.collectors {
		wg.Add(1)
		go func(col Collector) {
			defer wg.Done()

			ticker := time.NewTicker(col.Interval())
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					data, err := col.Collect(ctx)
					if err != nil {
						slog.Error("collect failed", "name", col.Name(), "error", err)
						continue
					}

					m.out <- ito.Telemetry{
						NodeID:    m.nodeID,
						Type:      col.Name(),
						Timestamp: time.Now().UnixNano(),
						Payload:   data,
					}
				}
			}
		}(c)
	}

	wg.Wait()
}
