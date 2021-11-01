package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestBackend_get(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	testLogger, hooks := test.NewNullLogger()
	SetLogger(testLogger)

	defer cleanup()

	case1Filename := "terraform_env"
	case1Content := "{\"content\": \"bla bla\"}"
	createFile(tmpTestDir, case1Filename+".tfstate", case1Content)
	case2Filename := "staging_other.tfstate"
	case2Content := "I don't care about the content"
	createFile(tmpTestDir, case2Filename, case2Content)

	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     []byte
		wantErr  bool
		wantLogs []string
	}{
		{
			"success get file content",
			fields{tmpTestDir},
			args{case1Filename},
			[]byte(case1Content),
			false,
			nil,
		},
		{
			"success get file content with given .tfstate extension",
			fields{tmpTestDir},
			args{case2Filename},
			[]byte(case2Content),
			false,
			nil,
		},
		{
			"try to get not existing file",
			fields{tmpTestDir},
			args{"this_file_not_exists"},
			nil,
			true,
			[]string{"this_file_not_exists.tfstate not found"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.get(tt.args.tfID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			checkLogMessage(t, tt.wantLogs, hooks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_getTfstateFilename(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"get file name without extension", fields{"/tmp/"}, args{"the_file"}, filepath.Join("/tmp/", "the_file.tfstate")},
		{"get file name with extension", fields{"/tmp/"}, args{"otherfile.tfstate"}, filepath.Join("/tmp/", "otherfile.tfstate")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got := b.getTfstateFilename(tt.args.tfID)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackend_lock(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	testLogger, hooks := test.NewNullLogger()
	SetLogger(testLogger)
	var lockInfo1, lockInfo2 LockInfo
	lockInfo1 = LockInfo{"myid1", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo1Bytes, _ := json.Marshal(lockInfo1)
	lockInfo2 = LockInfo{"myid2", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo2Bytes, _ := json.Marshal(lockInfo2)
	createFile(tmpTestDir, "exists_statelock.lock", string(lockInfo2Bytes))
	defer cleanup()
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
		lock []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     []byte
		wantErr  bool
		wantLogs []string
	}{
		{
			"set lock",
			fields{tmpTestDir},
			args{"this_state", lockInfo1Bytes},
			lockInfo1Bytes,
			false,
			nil,
		},
		{
			"update lock",
			fields{tmpTestDir},
			args{"this_state", lockInfo1Bytes},
			lockInfo1Bytes,
			false,
			nil,
		},
		{
			"trigger conflict",
			fields{tmpTestDir},
			args{"this_state", lockInfo2Bytes},
			nil,
			true,
			[]string{"state is locked with diffrend id myid1, but follow id requestd lock myid2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.lock(tt.args.tfID, tt.args.lock)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
			checkLogMessage(t, tt.wantLogs, hooks)
		})
	}
}

func TestBackend_purge(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	testLogger, hooks := test.NewNullLogger()
	SetLogger(testLogger)
	defer cleanup()

	case1TfStateFile := "existing.tfstate"
	createFile(tmpTestDir, case1TfStateFile, "some content")
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantLogs []string
	}{
		{"purge state", fields{tmpTestDir}, args{case1TfStateFile}, false, nil},
		{"ignore purge for not existing file", fields{tmpTestDir}, args{"not_existing_file"}, false, []string{"not_existing_file.tfstate not found"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.purge(tt.args.tfID); err != nil {
				assert.Error(t, err)
			}
			checkLogMessage(t, tt.wantLogs, hooks)
		})
	}
}

func TestBackend_unlock(t *testing.T) {
	var lockInfo1, lockInfo2 LockInfo
	tmpTestDir, cleanup := createDirectory()
	testLogger, hooks := test.NewNullLogger()
	SetLogger(testLogger)
	defer cleanup()

	lockInfo1 = LockInfo{"myid1", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo1Bytes, _ := json.Marshal(lockInfo1)
	lockInfo2 = LockInfo{"myid2", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo2Bytes, _ := json.Marshal(lockInfo2)

	createFile(tmpTestDir, "lockInfo1.lock", string(lockInfo1Bytes))
	createFile(tmpTestDir, "lockInfo2.lock", string(lockInfo2Bytes))

	type fields struct {
		dir string
	}
	type args struct {
		tfID string
		lock []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantLogs []string
	}{
		{"case1", fields{tmpTestDir}, args{"unlock_no_exists", lockInfo1Bytes}, false, []string{"unlock_no_exists.lock is deleted so notting to do."}},
		{"case2", fields{tmpTestDir}, args{"lockInfo1", lockInfo1Bytes}, false, nil},
		{"case3", fields{tmpTestDir}, args{"lockInfo2", lockInfo1Bytes}, true, []string{"state is locked with diffrend id myid2, but follow id requestd lock myid1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			err := b.unlock(tt.args.tfID, tt.args.lock)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			checkLogMessage(t, tt.wantLogs, hooks)
		})
	}
}

func TestBackend_update(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	defer cleanup()

	createFile(tmpTestDir, "oldfile.tfstate", "Old file content")

	type fields struct {
		dir string
	}
	type args struct {
		tfID    string
		tfstate []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Create File", fields{tmpTestDir}, args{"newfile", []byte("The file content")}, false},
		{"Create File", fields{tmpTestDir}, args{"oldfile", []byte("The file content")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			err := b.update(tt.args.tfID, tt.args.tfstate)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			fileContent, _ := os.ReadFile(b.getTfstateFilename(tt.args.tfID))
			assert.Equal(t, tt.args.tfstate, fileContent)
		})
	}
}

func TestConflictError_Error(t *testing.T) {
	type fields struct {
		StatusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"check error function", fields{1}, "status 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConflictError{
				StatusCode: tt.fields.StatusCode,
			}
			got := c.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileNotExistsError_Error(t *testing.T) {
	type fields struct {
		Info string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"TestFileNotExistsError", fields{"Information"}, "Information"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &FileNotExistsError{
				Info: tt.fields.Info,
			}
			got := n.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}
