package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	sshfsID       = "_sshfs"
	socketAddress = "/run/docker/plugins/sshfs.sock"
)

type sshfsVolume struct {
	password string
	sshcmd   string

	mountpoint  string
	connections int
}

type sshfsDriver struct {
	sync.RWMutex

	root    string
	volumes map[string]*sshfsVolume
}

func newSshfsDriver(root string) *sshfsDriver {
	d := &sshfsDriver{
		root:    root,
		volumes: make(map[string]*sshfsVolume),
	}

	return d
}

func (d *sshfsDriver) Create(r volume.Request) volume.Response {
	logrus.WithField("method", "create").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
	v := &sshfsVolume{}
	if r.Options == nil || r.Options["sshcmd"] == "" {
		return responseError("ssh option required")
	}
	v.sshcmd = r.Options["sshcmd"]
	v.password = r.Options["password"]
	v.mountpoint = filepath.Join(d.root, fmt.Sprintf("%x", md5.Sum([]byte(v.sshcmd))))

	d.volumes[r.Name] = v
	return volume.Response{}
}

func (d *sshfsDriver) Remove(r volume.Request) volume.Response {
	logrus.WithField("method", "remove").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return responseError(fmt.Sprintf("volume %s not found", r.Name))
	}

	if v.connections == 0 {
		if err := os.RemoveAll(v.mountpoint); err != nil {
			return responseError(err.Error())
		}
		delete(d.volumes, r.Name)
		return volume.Response{}
	}
	return responseError(fmt.Sprintf("volume %s is currently used by a container", r.Name))
}

func (d *sshfsDriver) Path(r volume.Request) volume.Response {
	logrus.WithField("method", "path").Debugf("%#v", r)

	d.RLock()
	defer d.RUnlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return responseError(fmt.Sprintf("volume %s not found", r.Name))
	}

	return volume.Response{Mountpoint: v.mountpoint}
}

func (d *sshfsDriver) Mount(r volume.MountRequest) volume.Response {
	logrus.WithField("method", "mount").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return responseError(fmt.Sprintf("volume %s not found", r.Name))
	}

	if v.connections > 0 {
		v.connections++
		return volume.Response{Mountpoint: v.mountpoint}
	}

	fi, err := os.Lstat(v.mountpoint)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(v.mountpoint, 0755); err != nil {
			return responseError(err.Error())
		}
	} else if err != nil {
		return responseError(err.Error())
	}

	if fi != nil && !fi.IsDir() {
		return responseError(fmt.Sprintf("%v already exist and it's not a directory", v.mountpoint))
	}

	if err := d.mountVolume(v); err != nil {
		return responseError(err.Error())
	}

	return volume.Response{Mountpoint: v.mountpoint}
}

func (d *sshfsDriver) Unmount(r volume.UnmountRequest) volume.Response {
	logrus.WithField("method", "unmount").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
	v, ok := d.volumes[r.Name]
	if !ok {
		return responseError(fmt.Sprintf("volume %s not found", r.Name))
	}
	if v.connections <= 1 {
		if err := d.unmountVolume(v.mountpoint); err != nil {
			return responseError(err.Error())
		}
		v.connections = 0
	} else {
		v.connections--
	}

	return volume.Response{}
}

func (d *sshfsDriver) Get(r volume.Request) volume.Response {
	logrus.WithField("method", "get").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return responseError(fmt.Sprintf("volume %s not found", r.Name))
	}

	return volume.Response{Volume: &volume.Volume{Name: r.Name, Mountpoint: v.mountpoint}}
}

func (d *sshfsDriver) List(r volume.Request) volume.Response {
	logrus.WithField("method", "list").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	var vols []*volume.Volume
	for name, v := range d.volumes {
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: v.mountpoint})
	}
	return volume.Response{Volumes: vols}
}

func (d *sshfsDriver) Capabilities(r volume.Request) volume.Response {
	logrus.WithField("method", "capabilities").Debugf("%#v", r)

	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func (d *sshfsDriver) mountVolume(v *sshfsVolume) error {
	cmd := fmt.Sprintf("sshfs -oStrictHostKeyChecking=no %s %s", v.sshcmd, v.mountpoint)
	if v.password != "" {
		cmd = fmt.Sprintf("echo %s | %s -o workaround=rename -o password_stdin", v.password, cmd)
	}
	logrus.Debug(cmd)
	return exec.Command("sh", "-c", cmd).Run()
}

func (d *sshfsDriver) unmountVolume(target string) error {
	cmd := fmt.Sprintf("umount %s", target)
	logrus.Debug(cmd)
	return exec.Command("sh", "-c", cmd).Run()
}

func responseError(err string) volume.Response {
	logrus.Error(err)
	return volume.Response{Err: err}
}

func main() {
	debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}

	d := newSshfsDriver(filepath.Join("/mnt", sshfsID))
	h := volume.NewHandler(d)
	logrus.Infof("listening on %s", socketAddress)
	logrus.Error(h.ServeUnix("", socketAddress))
}
