package util

import (
	"fmt"
	"sync/atomic"
	"time"
)

// NewTimeMark create TimeMark struct with time.Now start pos
func NewTimeMark() *TimeMark {
	return &TimeMark{
		start:   time.Now(),
		counter: 0,
	}
}

// TimeMark wrapper struct for time.Time
type TimeMark struct {
	start   time.Time
	counter int64
}

func (c *TimeMark) Counter() int64 {
	return atomic.LoadInt64(&c.counter)
}
func (c *TimeMark) Done(i int64) {
	atomic.AddInt64(&c.counter, i)
}

// CheckAndReset count time difference from time when TimeMark was created and time.Now then set create time to now
func (c *TimeMark) CheckAndReset() time.Duration {
	d := time.Since(c.start)
	c.start = time.Now()
	c.counter = 0
	return d
}

// Check count time difference from time when TimeMark was created and time.Now
func (c *TimeMark) Check() time.Duration {
	return time.Since(c.start)
}

// CountSIP counts time for one iteration
func (c *TimeMark) CountSIP() float32 {
	elapsed := c.Check()
	return float32(elapsed.Seconds()) / float32(c.counter)
}

// CountIPS counts how many iterations have been done from time when TimeMark was created
func (c *TimeMark) CountIPS() float32 {
	elapsed := c.Check()
	return float32(c.counter) / float32(elapsed.Seconds())
}

func (c *TimeMark) String() string {
	return fmt.Sprintf("time:%s IPS:%f SIP:%f", c.Check(), c.CountIPS(), c.CountSIP())
}
