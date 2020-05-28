package tester

import (
	"github.com/stretchr/testify/assert"
)

var NotNilValue = "not nil value"

type Case struct {
	In  []interface{}
	Out []interface{}
}

type TableTestCallback func(t *TableTestDriver, in []interface{}) []interface{}

type TableTestDriver struct {
	*assert.Assertions
	Cases []Case
}

func (t *TableTestDriver) Launch(cb TableTestCallback) {
	for i := 0; i < len(t.Cases); i++ {
		in := t.Cases[i].In
		expectedOut := t.Cases[i].Out
		realOut := cb(t, in)
		if len(expectedOut) != len(realOut) {
			t.Fail("len(expectedOut)!= len(realOut)")
		}
		for j := range expectedOut {
			e := expectedOut[j]
			r := realOut[j]
			switch {
			case e == NotNilValue:
				if r == nil {
					t.Fail("e == NotNilValue && r == nil˚")
				}
			default:
				t.Equal(e, r)
			}
		}
	}
}
