package collector

import (
	"context"
	"time"

	"github.com/orimono/ito"
)

type Collector interface {
	Name() string
	Interval() time.Duration
	Capability() ito.Capability
	Collect(ctx context.Context) (any, error)
}
