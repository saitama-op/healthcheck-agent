//go:build !linux

package checker

import "net"

// bindToInterface is a no-op on non-Linux systems, as SO_BINDTODEVICE is Linux-only.
func bindToInterface(dialer *net.Dialer, ifaceName string) {
	// macOS and Windows will rely entirely on LocalAddr IP binding
}
