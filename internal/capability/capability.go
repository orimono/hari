package capability

import (
	"context"
	"encoding/json"
	"time"

	"github.com/orimono/ito"
)

// Capable is the base interface shared by all capability-bearing components.
type Capable interface {
	Name() string
	Capability() ito.Capability
}

// Collector periodically gathers metrics.
type Collector interface {
	Capable
	Interval() time.Duration
	Collect(ctx context.Context) (any, error)
}

// Executor runs on-demand tasks dispatched by the server.
type Executor interface {
	Capable
	Execute(ctx context.Context, params json.RawMessage) (any, error)
}
