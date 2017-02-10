// +build !linux,!freebsd

package sdk

import (
	"errors"
	"net"
)

var (
	errOnlySupportedOnLinuxAndFreeBSD = errors.New("unix socket creation is only supported on linux and freebsd")
)

func newUnixListener(pluginName string, gid int) (net.Listener, string, error) {
	return nil, "", errOnlySupportedOnLinuxAndFreeBSD
}
