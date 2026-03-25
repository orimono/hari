package collector

import (
	"context"
	"time"
)

type Collector interface {
	Name() string
	Interval() time.Duration
	Collect(ctx context.Context) (any, error)
}
