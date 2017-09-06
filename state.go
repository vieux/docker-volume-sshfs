package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
)

func saveState(statePath string, volumes map[string]*sshfsVolume) error {
	data, err := json.Marshal(volumes)
	if err != nil {
		logrus.WithField("statePath", statePath).Error(err)
		return err
	}

	if err := ioutil.WriteFile(statePath, data, 0644); err != nil {
		logrus.WithField("savestate", statePath).Error(err)
		return err
	}
	return nil
}

func readState(statePath string) (map[string]*sshfsVolume, error) {
	volumes := make(map[string]*sshfsVolume)

	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("statePath", statePath).Debug("no state found")
		} else {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &volumes); err != nil {
			return nil, err
		}
	}
	return volumes, nil
}
