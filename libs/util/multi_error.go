package util

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
)

func NewMultiError() *MultiError {
	return &MultiError{
		errMap: make(map[string]error),
	}
}

type MultiError struct {
	errMap map[string]error
	mx     sync.RWMutex
}

func (m *MultiError) Add(e error, keys ...interface{}) {
	if e == nil {
		return
	}
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.errMap == nil {
		m.errMap = make(map[string]error)
	}

	if len(keys) > 0 {
		for _, key := range keys {
			m.errMap[fmt.Sprintf("%v", key)] = e
		}
	} else {
		counter := int64(len(m.errMap))
		key := strconv.FormatInt(counter, 10)
		m.errMap[key] = e
	}
}
func (m *MultiError) Error() string {
	m.mx.RLock()
	defer m.mx.RUnlock()
	var msg bytes.Buffer
	for errKey, err := range m.errMap {
		msg.WriteString(errKey)
		msg.WriteString(": ")
		msg.WriteString(err.Error())
		msg.WriteString(":\n")
	}
	return msg.String()
}
func (m *MultiError) CheckByKey(k interface{}) error {
	m.mx.RLock()
	defer m.mx.RUnlock()
	if m.errMap == nil {
		return nil
	}
	key := fmt.Sprintf("%v", k)
	e, ok := m.errMap[key]

	if ok {
		return e
	}
	return nil
}
func (m *MultiError) AnyNotNil() error {
	m.mx.RLock()
	defer m.mx.RUnlock()
	for _, err := range m.errMap {
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *MultiError) Check() error {
	m.mx.RLock()
	defer m.mx.RUnlock()
	n := len(m.errMap)
	if n > 0 {
		return m
	}
	return nil
}
