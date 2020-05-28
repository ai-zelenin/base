// Package rx contains types for reactive programming paradigm
package rx

import (
	"context"
	"sync"
)

// Callback function witch will be called on push
type Callback func(ctx context.Context, i interface{})

// NewNotifier create new Notifier struct
func NewNotifier(ctx context.Context) *Notifier {
	return &Notifier{
		ctx:         ctx,
		lock:        new(sync.RWMutex),
		callbackMap: make(map[string]Callback),
		observerMap: make(map[uint64]*Observer),
	}
}

// Notifier can register and remove callback functions witch will be called on each Notify
type Notifier struct {
	ctx         context.Context
	lock        *sync.RWMutex
	callbackMap map[string]Callback
	observerMap map[uint64]*Observer
}

// Notify exec all stored callbacks with defined argument. Concurrent safe.
func (n *Notifier) Notify(object interface{}) {
	n.lock.RLock()
	for _, callback := range n.callbackMap {
		if callback != nil {
			go func(ctx context.Context, obj interface{}, cb Callback) {
				cb(ctx, obj)
			}(n.ctx, object, callback)
		}
	}
	for _, o := range n.observerMap {
		o.Notify(object)
	}
	n.lock.RUnlock()
}

func (n *Notifier) CreateObserver(cb Callback, bufferSize int) *Observer {
	o := NewObserver(n.ctx, cb, bufferSize)
	n.lock.Lock()
	n.observerMap[o.id] = o
	n.lock.Unlock()
	go func() {
		<-o.ctx.Done()
		n.lock.Lock()
		delete(n.observerMap, o.id)
		n.lock.Unlock()
	}()
	return o
}

// Subscribe append new callback function with this name. Concurrent safe.
func (n *Notifier) Subscribe(name string, cb Callback) {
	n.lock.Lock()
	n.callbackMap[name] = cb
	n.lock.Unlock()
}

// Unsubscribe remove callback by name. Concurrent safe.
func (n *Notifier) Unsubscribe(name string) {
	n.lock.Lock()
	delete(n.callbackMap, name)
	n.lock.Unlock()
}
