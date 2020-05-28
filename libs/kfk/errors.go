package kfk

import (
	"fmt"
)

type ErrBrokerNotAvailable struct {
	Cause error
}

func (e ErrBrokerNotAvailable) Error() string {
	return fmt.Sprintf("cannot connect to broker: %v", e.Cause)
}
