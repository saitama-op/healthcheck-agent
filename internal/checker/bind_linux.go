//go:build linux

package checker

import (
	"net"
	"syscall"
)

// bindToInterface uses SO_BINDTODEVICE to force traffic out a physical interface,
// ignoring the default routing table.
func bindToInterface(dialer *net.Dialer, ifaceName string) {
	if ifaceName == "" {
		return
	}
	dialer.Control = func(network, address string, c syscall.RawConn) error {
		var err error
		controlErr := c.Control(func(fd uintptr) {
			err = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifaceName)
		})
		if controlErr != nil {
			return controlErr
		}
		return err
	}
}
