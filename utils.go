package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
)

func mountVolume(v *sshfsVolume) error {
	cmd := exec.Command("sshfs", "-oStrictHostKeyChecking=no", v.Sshcmd, v.Mountpoint)
	if v.Port != "" {
		cmd.Args = append(cmd.Args, "-p", v.Port)
	}
	if v.Password != "" {
		cmd.Args = append(cmd.Args, "-o", "workaround=rename", "-o", "password_stdin")
		cmd.Stdin = strings.NewReader(v.Password)
	}

	for _, option := range v.Options {
		cmd.Args = append(cmd.Args, "-o", option)
	}

	logrus.Debug(cmd.Args)
	return cmd.Run()
}

func unmountVolume(target string) error {
	cmd := fmt.Sprintf("umount %s", target)
	logrus.Debug(cmd)
	return exec.Command("sh", "-c", cmd).Run()
}

func logError(format string, args ...interface{}) error {
	logrus.Errorf(format, args...)
	return fmt.Errorf(format, args)
}
