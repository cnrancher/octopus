// +build !linux

package socket

import (
	"fmt"
	"io"
)

// NewSocket is a dummy function for non-Linux platform.
func NewSocket(id int) (io.ReadWriteCloser, error) {
	return nil, fmt.Errorf("only available on linux")
}
