package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ConflictError if there is a locking conflict
type ConflictError struct {
	StatusCode int
}

func (c *ConflictError) Error() string {
	return fmt.Sprintf("status %d", c.StatusCode)
}

// FileNotExistsError used if the file not exists
type FileNotExistsError struct {
	Info string
}

func (n *FileNotExistsError) Error() string {
	return n.Info
}

// LockInfo copied from https://github.com/hashicorp/terraform/blob/main/internal/states/statemgr/locker.go#L115
type LockInfo struct {
	ID        string
	Operation string
	Info      string
	Who       string
	Version   string
	Created   time.Time
	Path      string
}

// Backend struct to serve as backend for terraform http backend storage
type Backend struct {
	dir string
}

func (b *Backend) getTfstateFilename(tfID string) string {
	if strings.HasSuffix(tfID, ".tfstate") {
		return b.dir + tfID
	}
	return b.dir + tfID + ".tfstate"
}

func (b *Backend) get(tfID string) ([]byte, error) {
	var tfstateFilename = b.getTfstateFilename(tfID)
	var tfstate []byte
	var err error

	if _, err := os.Stat(tfstateFilename); os.IsNotExist(err) {
		log.Infof("File %s not found", tfstateFilename)
		return nil, err
	}
	if tfstate, err = ioutil.ReadFile(tfstateFilename); err != nil {
		log.Warnf("Can't read file %s. With follow error %v", tfstateFilename, err)
		return nil, err
	}

	return tfstate, nil
}

func (b *Backend) update(tfID string, tfstate []byte) error {
	var tfstateFilename = b.getTfstateFilename(tfID)

	if err := ioutil.WriteFile(tfstateFilename, tfstate, 0644); err != nil {
		log.Warnf("Can't write file %s. Got follow error %v", tfstateFilename, err)
		return err
	}
	return nil
}

func (b *Backend) pruge(tfID string) error {
	var tfstateFilename = b.getTfstateFilename(tfID)

	if _, err := os.Stat(tfstateFilename); os.IsNotExist(err) {
		log.Infof("File %s not found", tfstateFilename)
		return nil
	}

	if err := os.Remove(tfstateFilename); err != nil {
		log.Warnf("Can't delete file %s. Got follow error %v", tfstateFilename, err)
		return err
	}
	return nil
}

func (b *Backend) lock(tfID string, lock []byte) ([]byte, error) {
	var lockFilename string = b.dir + tfID + ".lock"
	var lockFile []byte
	var lockInfo, currentLockInfo LockInfo
	var err error

	if err := json.Unmarshal(lock, &lockInfo); err != nil {
		log.Errorf("unexpected decoding json error %v", err)
		return nil, err
	}
	if _, err := os.Stat(lockFilename); os.IsNotExist(err) {
		if lockFile, err = json.MarshalIndent(lockInfo, "", " "); err != nil {
			log.Errorf("unexpected encoding json error %v", err)
			return nil, err
		}
		if err = ioutil.WriteFile(lockFilename, lockFile, 0644); err != nil {
			log.Errorf("Can't write lock file %s. Got follow error %v", lockFilename, err)
			return nil, err
		}
		return lockFile, nil
	}

	if lockFile, err = ioutil.ReadFile(lockFilename); err != nil {
		log.Errorf("Can't read file %s. With follow error %v", lockFilename, err)
		return nil, err
	}
	if err := json.Unmarshal(lockFile, &currentLockInfo); err != nil {
		log.Errorf("unexpected decoding json error %v", err)
		return nil, err
	}
	if currentLockInfo.ID != lockInfo.ID {
		log.Infof("state is locked with diffrend id %s, but follow id requestd lock %s", currentLockInfo.ID, lockInfo.ID)
		return nil, &ConflictError{
			StatusCode: http.StatusConflict,
		}
	}
	return lockFile, nil
}

func (b *Backend) unlock(tfID string, lock []byte) error {
	var lockFilename string = b.dir + tfID + ".lock"
	var lockFile []byte
	var err error
	var lockInfo, currentLockInfo LockInfo

	if err := json.Unmarshal(lock, &lockInfo); err != nil {
		log.Errorf("unexpected decoding json error %v", err)
		return err
	}
	if _, err := os.Stat(lockFilename); os.IsNotExist(err) {
		log.Infof("lock file %s is deleted so notting to do.", lockFilename)
		return nil
	}
	if lockFile, err = ioutil.ReadFile(lockFilename); err != nil {
		log.Errorf("Can't read file %s. With follow error %v", lockFilename, err)
		return err
	}
	if err := json.Unmarshal(lockFile, &currentLockInfo); err != nil {
		log.Errorf("unexpected decoding json error %v", err)
		return err
	}
	if currentLockInfo.ID != lockInfo.ID {
		log.Infof("state is locked with diffrend id %s, but follow id requestd lock %s", currentLockInfo.ID, lockInfo.ID)
		return &ConflictError{
			StatusCode: http.StatusConflict,
		}
	}
	if err := os.Remove(lockFilename); err != nil {
		log.Warnf("Can't delete file %s. Got follow error %v", lockFilename, err)
		return err
	}

	return nil
}
