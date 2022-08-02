package thirdparty

import (
	"fmt"
)

type Isolation int

const (
	IsolationChroot      Isolation = 2
	IsolationOCIRootless Isolation = 3
)

func (i Isolation) String() string {
	switch i {
	case IsolationChroot:
		return "chroot"
	case IsolationOCIRootless:
		return "rootless"
	}
	return fmt.Sprintf("unrecognized isolation type %d", i)
}
