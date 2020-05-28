package util

import (
	"sync"
	"sync/atomic"
)

// NewMultiLock create a new multiple lock
func NewMultiLock() *MultiLock {
	return &MultiLock{
		pool: &sync.Pool{
			New: func() interface{} {
				return &sync.RWMutex{}
			},
		},
	}
}

// MultiLock for any keys
type MultiLock struct {
	inUse sync.Map
	pool  *sync.Pool
}
type refCounter struct {
	counter int64
	lock    *sync.RWMutex
}

func (l *MultiLock) Lock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.Lock()
}

func (l *MultiLock) RLock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.RLock()
}

func (l *MultiLock) Unlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.Unlock()
	l.putBackInPool(key, m)
}

func (l *MultiLock) RUnlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.RUnlock()
	l.putBackInPool(key, m)
}

func (l *MultiLock) putBackInPool(key interface{}, m *refCounter) {
	atomic.AddInt64(&m.counter, -1)
	if m.counter <= 0 {
		l.pool.Put(m.lock)
		l.inUse.Delete(key)
	}
}

func (l *MultiLock) getLocker(key interface{}) *refCounter {
	res, _ := l.inUse.LoadOrStore(key, &refCounter{
		counter: 0,
		lock:    l.pool.Get().(*sync.RWMutex),
	})

	return res.(*refCounter)
}
