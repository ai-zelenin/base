package util

import (
	"fmt"
	"time"
)

// TimeGridEvent struct with scale and time
type TimeGridEvent struct {
	Scale string
	Time  time.Time
}

// CreateTimeGridEmitter create time grid for every scale in list find nears time in grid and start ticker in this time
func CreateTimeGridEmitter(channel chan TimeGridEvent, scaleList ...string) {
	if channel == nil {
		panic(fmt.Errorf("channel TimeGridEvent is nil"))
	}
	for _, scale := range scaleList {
		startSignal := make(chan time.Duration, 1)
		duration, err := time.ParseDuration(scale)
		if err != nil {
			panic(err)
		}
		now := time.Now()
		day := time.Hour * 24
		// First cell in time grid
		t0 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		// Last cell in time grid
		tn := t0.Add(day)
		// Number of cells in time grid
		n := int(day / duration)

		scaleTimeGrid := make([]time.Time, n+1)
		scaleTimeGrid[0] = t0
		scaleTimeGrid[len(scaleTimeGrid)-1] = tn

		for i := 0; i < n; i++ {
			nextTime := scaleTimeGrid[i].Add(duration)
			if nextTime != tn {
				scaleTimeGrid[i+1] = nextTime
			}
		}
		scaleTimeGrid = scaleTimeGrid[1:]

		go func(s string) {
			d, ok := <-startSignal
			if ok {
				ticker := time.NewTicker(d)
				for tick := range ticker.C {
					channel <- TimeGridEvent{Scale: s, Time: tick}
				}
			}
		}(scale)

		var startDelay time.Duration
		for _, t := range scaleTimeGrid {
			diff := time.Until(t)
			if diff > 0 {
				startDelay = diff
				break
			}
		}
		time.AfterFunc(startDelay, func() {
			startSignal <- duration
			close(startSignal)
		})
	}
}
