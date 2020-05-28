package tester

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/mock"
)

func MonkeyMock(obj interface{}, methodName string, args ...interface{}) *MonkeyPatcher {
	mp := &MonkeyPatcher{
		ExpectedCalls:      make(map[string]bool),
		ExpectedArgObjects: args,
		Target:             obj,
		MethodName:         methodName,
	}
	mp.ExpectedCalls[mp.formatCall(args)] = false
	return mp
}

type MonkeyPatcher struct {
	ExpectedCalls       map[string]bool
	Target              interface{}
	MethodName          string
	ExpectedArgObjects  []interface{}
	ReturnObjects       []interface{}
	ReplacementFunction interface{}
	Guard               *monkey.PatchGuard
}

func (m *MonkeyPatcher) Return(r ...interface{}) *MonkeyPatcher {
	m.ReturnObjects = r
	target := reflect.TypeOf(m.Target)
	method, ok := target.MethodByName(m.MethodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", m.MethodName))
	}

	if m.ReplacementFunction == nil {
		f := m.makeFunction(method)
		m.ReplacementFunction = f
	}

	m.Guard = monkey.PatchInstanceMethod(target, m.MethodName, m.ReplacementFunction)
	return m
}

func (m *MonkeyPatcher) ReplaceFunc(f interface{}) *MonkeyPatcher {
	m.ReplacementFunction = f
	target := reflect.TypeOf(m.Target)
	m.Guard = monkey.PatchInstanceMethod(target, m.MethodName, m.ReplacementFunction)
	return m
}

func (m *MonkeyPatcher) Unpatch() {
	m.Guard.Unpatch()
}

func (m *MonkeyPatcher) makeFunction(method reflect.Method) interface{} {
	outTypes := make([]reflect.Type, 0)
	for i := 0; i < method.Func.Type().NumOut(); i++ {
		outTypes = append(outTypes, method.Func.Type().Out(i))
	}
	if method.Func.Type().NumIn() != len(m.ExpectedArgObjects)+1 {
		panic("method.Func.Type().NumIn() != len(m.ExpectedArgObjects)")
	}
	if method.Func.Type().NumOut() != len(m.ReturnObjects) {
		panic("method.Func.Type().NumOut() != len(m.ReturnObjects)")
	}
	fOut := func(args []reflect.Value) []reflect.Value {
		castedArgs := make([]interface{}, 0, len(args))
		for _, arg := range args[1:] {
			castedArgs = append(castedArgs, arg.Interface())
		}

		isExpected := m.isExpected(castedArgs)
		callSignature := m.formatCall(castedArgs)
		var result []reflect.Value
		for i, ro := range m.ReturnObjects {
			if ro != nil {
				result = append(result, reflect.ValueOf(ro))
			} else {
				result = append(result, reflect.Zero(outTypes[i]))
			}
		}
		if isExpected {
			_, ok := m.ExpectedCalls[callSignature]
			if ok {
				m.ExpectedCalls[callSignature] = true
			}
		} else {
			m.ExpectedCalls[callSignature] = false
		}

		return result
	}
	return reflect.MakeFunc(reflect.TypeOf(method.Func.Interface()), fOut).Interface()
}

func (m *MonkeyPatcher) AssertExpectations(t *testing.T) {
	for k, v := range m.ExpectedCalls {
		if !v {
			t.Error(fmt.Errorf("bad call: %s", k))
			return
		}
	}
}

func (m *MonkeyPatcher) formatCall(args []interface{}) string {
	return fmt.Sprintf("MonkeyMockCall:[%s]-%p: args:%v", m.MethodName, m, args)
}

func (m *MonkeyPatcher) isExpected(args []interface{}) bool {
	for i, arg := range args {
		expectedArg := m.ExpectedArgObjects[i]
		if expectedArg == mock.Anything {
			return true
		}
		if !reflect.DeepEqual(expectedArg, arg) {
			return false
		}
	}
	return true
}
