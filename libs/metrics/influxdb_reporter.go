package metrics

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/uber-go/tally"

	"git.pnhub.ru/core/libs/influx"
	"git.pnhub.ru/core/libs/log"
)

func NewInfluxDBReporter(ctx context.Context, logger log.Logger, influxCli *influx.Client) (tally.StatsReporter, error) {
	if influxCli == nil {
		return nil, fmt.Errorf("no influx client ")
	}
	return &InfluxDBReporter{
		ctx:    ctx,
		logger: log.ForkLogger(logger),
		cli:    influxCli,
		points: make([]influx.Point, 0),
		mx:     new(sync.RWMutex),
	}, nil
}

type InfluxDBReporter struct {
	cli    *influx.Client
	ctx    context.Context
	logger log.Logger
	points []influx.Point
	mx     *sync.RWMutex
}

func (r *InfluxDBReporter) Reporting() bool {
	return true
}

func (r *InfluxDBReporter) Tagging() bool {
	return true
}

func (r *InfluxDBReporter) Capabilities() tally.Capabilities {
	return r
}

func (r *InfluxDBReporter) Flush() {
	r.mx.Lock()
	snap := make([]influx.Point, len(r.points))
	copy(snap, r.points)
	r.points = make([]influx.Point, 0)
	r.mx.Unlock()

	err := r.cli.WritePoints(snap)
	if err != nil {
		r.logger.Error(err)
	}
}

func (r *InfluxDBReporter) ReportCounter(name string, tags map[string]string, value int64) {
	r.addPoint(influx.Point{
		Time: time.Now(),
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"counter": float64(value),
		},
	})
}

func (r *InfluxDBReporter) ReportGauge(name string, tags map[string]string, value float64) {
	r.addPoint(influx.Point{
		Time: time.Now(),
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"gauge": value,
		},
	})
}

// ReportTimer a timer value
func (r *InfluxDBReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	r.addPoint(influx.Point{
		Time: time.Now(),
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"timing": float64(interval.Nanoseconds() / int64(time.Millisecond)),
		},
	})
}

// ReportHistogramValueSamples histogram samples for a bucket
func (r *InfluxDBReporter) ReportHistogramValueSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound float64, samples int64) {
	var snap = make(map[string]string, len(tags))
	for k, v := range tags {
		snap[k] = v
	}

	var timing = int64(bucketUpperBound)

	tags["timing_tag"] = strconv.FormatInt(timing, 10)

	point := influx.Point{
		Time: time.Now(),
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"timing": bucketUpperBound,
			"count":  float64(samples),
		},
	}

	r.addPoint(point)
}

// ReportHistogramDurationSamples histogram samples for a bucket
func (r *InfluxDBReporter) ReportHistogramDurationSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound time.Duration, samples int64) {
	var snap = make(map[string]string, len(tags))
	for k, v := range tags {
		snap[k] = v
	}
	var timing = bucketUpperBound.Nanoseconds() / int64(time.Millisecond)

	tags["timing_tag"] = strconv.FormatInt(timing, 10)

	point := influx.Point{
		Time: time.Now(),
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"timing": bucketUpperBound,
			"count":  float64(samples),
		},
	}

	r.addPoint(point)
}

func (r *InfluxDBReporter) addPoint(point influx.Point) {
	r.mx.Lock()
	r.points = append(r.points, point)
	r.mx.Unlock()
}
