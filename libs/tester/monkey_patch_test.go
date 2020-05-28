package tester

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMonkeyPatchSuite(t *testing.T) {
	suite.Run(t, new(MonkeyPatchSuite))
}

type MonkeyPatchSuite struct {
	suite.Suite
}

func (s *MonkeyPatchSuite) SetupSuite() {

}

type A struct {
	val int
}

func (a *A) Val() int {
	return a.val
}

func (s *MonkeyPatchSuite) TestMonkeyPatch() {
	a := &A{val: 1}

	MonkeyMock(a, "Val").ReplaceFunc(func(a *A) int {
		return 2
	})

	if a.Val() != 2 {
		s.Fail("monkey patch do not working")
	}

}
