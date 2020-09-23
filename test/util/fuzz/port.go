// +build test

package fuzz

import (
	"errors"
	"fmt"
	"net"

	"k8s.io/apimachinery/pkg/util/sets"
)

// FreePorts tries to find the free ports and returns them or error.
func FreePorts(amount int) ([]int, error) {
	var set = sets.NewInt()

	for {
		if set.Len() >= amount {
			break
		}
		var lis, err = net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("failed to get free ports, %v", err)
		}

		var addr, ok = lis.Addr().(*net.TCPAddr)
		if !ok {
			_ = lis.Close()
			return nil, errors.New("failed to get a TCP address")
		}
		set.Insert(addr.Port)
		_ = lis.Close()
	}

	return set.List(), nil
}

// FreePort tries to find a free port and returns them or error.
func FreePort() (int, error) {
	var ports, err = FreePorts(1)
	if err != nil {
		return 0, err
	}
	return ports[0], nil
}
