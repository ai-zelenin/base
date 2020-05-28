// Package metrics contains objects witch are using for metrics composition and transfer
package metrics

import (
	"time"

	"github.com/uber-go/tally"
)

func DefaultBuckets() tally.Buckets {
	return tally.MustMakeExponentialDurationBuckets(time.Millisecond*2, 2, 16)
}
