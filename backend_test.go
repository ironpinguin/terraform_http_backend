package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestBackend_get(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()

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
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"success get file content",
			fields{tmpTestDir},
			args{case1Filename},
			[]byte(case1Content),
			false,
		},
		{
			"success get file content with given .tfstate extension",
			fields{tmpTestDir},
			args{case2Filename},
			[]byte(case2Content),
			false,
		},
		{
			"try to get not existing file",
			fields{tmpTestDir},
			args{"this_file_not_exists"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.get(tt.args.tfID)
			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() got = %v, want %v", got, tt.want)
			}
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
		{"get file name without extension", fields{"/tmp/"}, args{"the_file"}, "/tmp/the_file.tfstate"},
		{"get file name with extension", fields{"/tmp/"}, args{"otherfile.tfstate"}, "/tmp/otherfile.tfstate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if got := b.getTfstateFilename(tt.args.tfID); got != tt.want {
				t.Errorf("getTfstateFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_lock(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
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
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"set lock",
			fields{tmpTestDir},
			args{"this_state", lockInfo1Bytes},
			lockInfo1Bytes,
			false,
		},
		{
			"update lock",
			fields{tmpTestDir},
			args{"this_state", lockInfo1Bytes},
			lockInfo1Bytes,
			false,
		},
		{
			"trigger conflict",
			fields{tmpTestDir},
			args{"this_state", lockInfo2Bytes},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.lock(tt.args.tfID, tt.args.lock)
			if (err != nil) != tt.wantErr {
				t.Errorf("lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_purge(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"purge state", fields{tmpTestDir}, args{case1TfStateFile}, false},
		{"ignore purge for not existing file", fields{tmpTestDir}, args{"not_existing_file"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.purge(tt.args.tfID); (err != nil) != tt.wantErr {
				t.Errorf("pruge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackend_unlock(t *testing.T) {
	var lockInfo1, lockInfo2 LockInfo
	tmpTestDir, cleanup := createDirectory()
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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"case1", fields{tmpTestDir}, args{"unlock_no_exists", lockInfo1Bytes}, false},
		{"case2", fields{tmpTestDir}, args{"lockInfo1", lockInfo1Bytes}, false},
		{"case3", fields{tmpTestDir}, args{"lockInfo2", lockInfo1Bytes}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.unlock(tt.args.tfID, tt.args.lock); (err != nil) != tt.wantErr {
				t.Errorf("unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			if err := b.update(tt.args.tfID, tt.args.tfstate); (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
			}
			fileContent, _ := os.ReadFile(b.getTfstateFilename(tt.args.tfID))
			if !bytes.Equal(tt.args.tfstate, fileContent) {
				t.Errorf("update() got %s want %s", string(fileContent), string(tt.args.tfstate))
			}
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
			if got := c.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
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
			if got := n.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
