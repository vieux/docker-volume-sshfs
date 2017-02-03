package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const socketAddress = "/run/docker/plugins/sshfs.sock"

type sshfsVolume struct {
	Password string
	Sshcmd   string

	Mountpoint  string
	connections int
}

type sshfsDriver struct {
	sync.RWMutex

	root      string
	statePath string
	volumes   map[string]*sshfsVolume
}

func newSshfsDriver(root string) (*sshfsDriver, error) {
	logrus.WithField("method", "new driver").Debug(root)

	d := &sshfsDriver{
		root:      filepath.Join(root, "volumes"),
		statePath: filepath.Join(root, "state", "sshfs-state.json"),
		volumes:   map[string]*sshfsVolume{},
	}

	data, err := ioutil.ReadFile(d.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("statePath", d.statePath).Debug("no state found")
		} else {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &d.volumes); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (d *sshfsDriver) saveState() {
	data, err := json.Marshal(d.volumes)
	if err != nil {
		logrus.WithField("statePath", d.statePath).Error(err)
		return
	}

	if err := ioutil.WriteFile(d.statePath, data, 0644); err != nil {
		logrus.WithField("savestate", d.statePath).Error(err)
	}
}

func (d *sshfsDriver) Create(r volume.Request) volume.Response {
	logrus.WithField("method", "create").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
	v := &sshfsVolume{}
	if r.Options == nil || r.Options["sshcmd"] == "" {
		return responseError("ssh option required")
	}
	v.Sshcmd = r.Options["sshcmd"]
	v.Password = r.Options["password"]
	v.Mountpoint = filepath.Join(d.root, fmt.Sprintf("%x", md5.Sum([]byte(v.Sshcmd))))

	d.volumes[r.Name] = v

	d.saveState()

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
		if err := os.RemoveAll(v.Mountpoint); err != nil {
			return responseError(err.Error())
		}
		delete(d.volumes, r.Name)
		d.saveState()
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

	return volume.Response{Mountpoint: v.Mountpoint}
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
		return volume.Response{Mountpoint: v.Mountpoint}
	}

	fi, err := os.Lstat(v.Mountpoint)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(v.Mountpoint, 0755); err != nil {
			return responseError(err.Error())
		}
	} else if err != nil {
		return responseError(err.Error())
	}

	if fi != nil && !fi.IsDir() {
		return responseError(fmt.Sprintf("%v already exist and it's not a directory", v.Mountpoint))
	}

	if err := d.mountVolume(v); err != nil {
		return responseError(err.Error())
	}

	return volume.Response{Mountpoint: v.Mountpoint}
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
		if err := d.unmountVolume(v.Mountpoint); err != nil {
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

	return volume.Response{Volume: &volume.Volume{Name: r.Name, Mountpoint: v.Mountpoint}}
}

func (d *sshfsDriver) List(r volume.Request) volume.Response {
	logrus.WithField("method", "list").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	var vols []*volume.Volume
	for name, v := range d.volumes {
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: v.Mountpoint})
	}
	return volume.Response{Volumes: vols}
}

func (d *sshfsDriver) Capabilities(r volume.Request) volume.Response {
	logrus.WithField("method", "capabilities").Debugf("%#v", r)

	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func (d *sshfsDriver) mountVolume(v *sshfsVolume) error {
	cmd := fmt.Sprintf("sshfs -oStrictHostKeyChecking=no %s %s", v.Sshcmd, v.Mountpoint)
	if v.Password != "" {
		cmd = fmt.Sprintf("echo %s | %s -o workaround=rename -o password_stdin", v.Password, cmd)
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

	d, err := newSshfsDriver("/mnt")
	if err != nil {
		log.Fatal(err)
	}
	h := volume.NewHandler(d)
	logrus.Infof("listening on %s", socketAddress)
	logrus.Error(h.ServeUnix("", socketAddress))
}
