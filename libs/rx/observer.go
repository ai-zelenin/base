package rx

import (
	"context"
	"math/rand"
)

// Observer type represent structure witch can react on events
type Observer struct {
	id     uint64
	ctx    context.Context
	cancel context.CancelFunc
	cb     Callback
	events chan interface{}
}

func NewObserver(ctx context.Context, cb Callback, bufferSize int) *Observer {
	oCtx, cancel := context.WithCancel(ctx)
	o := &Observer{
		ctx:    oCtx,
		cancel: cancel,
		cb:     cb,
		events: make(chan interface{}, bufferSize),
		id:     rand.Uint64(),
	}
	go o.observe()
	return o
}

func (o *Observer) observe() {
	for {
		select {
		case <-o.ctx.Done():
			return
		case obj, ok := <-o.events:
			if !ok {
				return
			}
			if o.cb != nil {
				o.cb(o.ctx, obj)
			}
		}
	}
}

func (o *Observer) Notify(obj interface{}) {
	o.events <- obj
}

func (o *Observer) Stop() {
	o.cb = nil
	o.cancel()
}
