package main

import (
	"crypto/md5"
	"fmt"
	"path/filepath"

	"github.com/Sirupsen/logrus"
)

type sshfsVolume struct {
	Password string
	Sshcmd   string
	Port     string

	Options []string

	Mountpoint  string
	connections int
}

func newSshfsVolume(root string, options map[string]string) (*sshfsVolume, error) {
	logrus.WithField("method", "new volume").Debugf("%s, %#v", root, options)

	var v sshfsVolume

	for key, val := range options {
		switch key {
		case "sshcmd":
			v.Sshcmd = val
		case "password":
			v.Password = val
		case "port":
			v.Port = val
		default:
			if val != "" {
				v.Options = append(v.Options, key+"="+val)
			} else {
				v.Options = append(v.Options, key)
			}
		}
	}

	if v.Sshcmd == "" {
		return nil, logError("'sshcmd' option required")
	}

	v.Mountpoint = filepath.Join(root, fmt.Sprintf("%x", md5.Sum([]byte(v.Sshcmd))))

	return &v, nil
}
