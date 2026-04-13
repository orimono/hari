package capability

import (
	"context"
	"log/slog"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/orimono/ito"
)

type Manager struct {
	nodeID     string
	collectors []Collector
	executors  []Executor
	out        chan ito.Telemetry
}

func NewManager(nodeID string) *Manager {
	return &Manager{
		nodeID:     nodeID,
		collectors: make([]Collector, 0),
		executors:  make([]Executor, 0),
		out:        make(chan ito.Telemetry, 100),
	}
}

func (m *Manager) AddCollector(c Collector) {
	platforms := c.Capability().Platforms
	if len(platforms) > 0 {
		if !slices.Contains(platforms, runtime.GOOS) {
			slog.Info("skipping collector: unsupported platform", "name", c.Name(), "platform", runtime.GOOS)
			return
		}
	}
	m.collectors = append(m.collectors, c)
}

func (m *Manager) Out() <-chan ito.Telemetry {
	return m.out
}

func (m *Manager) AddExecutor(e Executor) {
	platforms := e.Capability().Platforms
	if len(platforms) > 0 && !slices.Contains(platforms, runtime.GOOS) {
		slog.Info("skipping executor: unsupported platform", "name", e.Name(), "platform", runtime.GOOS)
		return
	}
	m.executors = append(m.executors, e)
}

func (m *Manager) Executors() []Executor {
	return m.executors
}

// Register adds an executor at runtime (e.g. a dynamically loaded script executor).
func (m *Manager) Register(e Executor) {
	m.executors = append(m.executors, e)
}

func (m *Manager) Manifest() map[string]ito.Capability {
	manifest := make(map[string]ito.Capability, len(m.collectors)+len(m.executors))
	for _, c := range m.collectors {
		manifest[c.Name()] = c.Capability()
	}
	for _, e := range m.executors {
		manifest[e.Name()] = e.Capability()
	}
	return manifest
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
