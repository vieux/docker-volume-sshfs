// +build linux freebsd

package sdk

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/util"
	"github.com/docker/go-connections/sockets"
)

const (
	pluginSockDir = "/run/docker/plugins"
)

func newUnixListener(pluginName string, group string) (net.Listener, string, error) {
	path, err := fullSocketAddress(pluginName)
	if err != nil {
		return nil, "", err
	}
	listener, err := setupSocketActivation()
	if err != nil {
		return nil, "", err
	}
	if listener == nil {
		listener, err = sockets.NewUnixSocket(path, group)
		if err != nil {
			return nil, "", err
		}
	}
	return listener, path, nil
}

func fullSocketAddress(address string) (string, error) {
	if err := os.MkdirAll(pluginSockDir, 0755); err != nil {
		return "", err
	}
	if filepath.IsAbs(address) {
		return address, nil
	}
	return filepath.Join(pluginSockDir, address+".sock"), nil
}

func setupSocketActivation() (net.Listener, error) {
	if !util.IsRunningSystemd() {
		return nil, nil
	}
	listenFds := activation.Files(false)
	if len(listenFds) > 1 {
		return nil, fmt.Errorf("expected only one socket from systemd, got %d", len(listenFds))
	}
	var listener net.Listener
	if len(listenFds) == 1 {
		l, err := net.FileListener(listenFds[0])
		if err != nil {
			return nil, err
		}
		listener = l
	}
	return listener, nil
}
