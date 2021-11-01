package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var hooks *test.Hook
var testLogger *logrus.Logger

func init() {
	testLogger, hooks = test.NewNullLogger()
	SetLogger(testLogger)
	config.loadConfig(".env.test")
}

func Test_getTfstate(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	defer cleanup()

	createFile(tmpTestDir, "state_file1.tfstate", "This is the content")

	type args struct {
		suburl string
	}
	tests := []struct {
		name string
		args
		wantStatus int
		wantBody   string
		wantLogs   []string
	}{
		{"First case", args{"state_file1"}, 200, "This is the content", nil},
		{"request unknown file", args{"no_file1"}, 404, "Not Found", []string{"no_file1.tfstate not found"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageBackend = Backend{tmpTestDir}
			router := chi.NewRouter()
			router.Get("/{id}", getTfstate)
			ts := httptest.NewServer(router)
			defer ts.Close()

			rr, got := testRequest(t, ts, "GET", "/"+tt.args.suburl, nil)

			assert.Equal(t, tt.wantStatus, rr.StatusCode)
			assert.Equal(t, tt.wantBody, got)
			checkLogMessage(t, tt.wantLogs)
		})
	}
}

func Test_lockTfstate(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	defer cleanup()

	var lockInfo1, lockInfo2 LockInfo
	lockInfo1 = LockInfo{"myid1", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo1Bytes, _ := json.Marshal(lockInfo1)
	lockInfo2 = LockInfo{"myid2", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo2Bytes, _ := json.Marshal(lockInfo2)
	createFile(tmpTestDir, "exists_statelock.lock", string(lockInfo2Bytes))

	type args struct {
		body   string
		suburl string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		wantLogs   []string
	}{
		{"Defect lock json body", args{"jfkdslf", "not_file"}, 500, "Internal Server Error", []string{"unexpected decoding json error"}},
		{"create new lock", args{string(lockInfo1Bytes), "not_file"}, 200, string(lockInfo1Bytes), nil},
		{"conflict with existing lock", args{string(lockInfo1Bytes), "exists_statelock"}, 409, "Conflict", []string{"state is locked with diffrend id myid2, but follow id requestd lock myid1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageBackend = Backend{tmpTestDir}
			router := chi.NewRouter()
			chi.RegisterMethod("LOCK")
			router.MethodFunc("LOCK", "/{id}", lockTfstate)
			ts := httptest.NewServer(router)
			defer ts.Close()

			rr, got := testRequest(t, ts, "LOCK", "/"+tt.args.suburl, strings.NewReader(tt.args.body))
			assert.Equal(t, tt.wantStatus, rr.StatusCode)
			assert.Equal(t, tt.wantBody, got)
			checkLogMessage(t, tt.wantLogs)
		})
	}
}

func Test_purgeTfstate(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	defer cleanup()

	createFile(tmpTestDir, "existing_state.tfstate", "This is the Content")

	type args struct {
		suburl string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		wantLogs   []string
	}{
		{
			"Delete existing file",
			args{"existing_state"},
			200,
			`{"state": "tfstate deleted"}`,
			nil,
		},
		{
			"Delete not existing file",
			args{"no_file_exists"},
			200,
			`{"state": "tfstate deleted"}`,
			[]string{"no_file_exists.tfstate not found"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageBackend = Backend{tmpTestDir}
			router := chi.NewRouter()
			router.Delete("/{id}", purgeTfstate)
			ts := httptest.NewServer(router)
			defer ts.Close()

			rr, got := testRequest(t, ts, "DELETE", "/"+tt.args.suburl, nil)
			assert.Equal(t, tt.wantStatus, rr.StatusCode)
			assert.Equal(t, tt.wantBody, got)
			checkLogMessage(t, tt.wantLogs)
		})
	}
}

func Test_unlockTfstate(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()

	defer cleanup()

	var lockInfo1, lockInfo2 LockInfo
	lockInfo1 = LockInfo{"myid1", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo1Bytes, _ := json.Marshal(lockInfo1)
	lockInfo2 = LockInfo{"myid2", "START", "ThisInfo", "", "", time.Now(), ""}
	lockInfo2Bytes, _ := json.Marshal(lockInfo2)
	createFile(tmpTestDir, "exists_statelock.lock", string(lockInfo2Bytes))

	type args struct {
		suburl string
		body   string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		wantLogs   []string
	}{
		{
			"Defect lock json body",
			args{"not_file", "jfkdslf"},
			500,
			"Internal Server Error",
			[]string{"unexpected decoding json error"},
		},
		{
			"unlock success on no file",
			args{"not_file", string(lockInfo1Bytes)},
			200,
			string(lockInfo1Bytes),
			[]string{"not_file.lock is deleted so notting to do"},
		},
		{
			"conflict with existing lock",
			args{"exists_statelock", string(lockInfo1Bytes)},
			409,
			"Conflict",
			[]string{"state is locked with diffrend id myid2, but follow id requestd lock myid1"},
		},
		{
			"unlock success on existing lock",
			args{"exists_statelock", string(lockInfo2Bytes)},
			200,
			string(lockInfo2Bytes),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageBackend = Backend{tmpTestDir}
			router := chi.NewRouter()
			chi.RegisterMethod("UNLOCK")
			router.MethodFunc("UNLOCK", "/{id}", unlockTfstate)
			ts := httptest.NewServer(router)
			defer ts.Close()

			rr, got := testRequest(t, ts, "UNLOCK", "/"+tt.args.suburl, strings.NewReader(tt.args.body))
			assert.Equal(t, tt.wantStatus, rr.StatusCode)
			assert.Equal(t, tt.wantBody, got)
			checkLogMessage(t, tt.wantLogs)
		})
	}
}

func Test_updateTfstate(t *testing.T) {
	tmpTestDir, cleanup := createDirectory()
	defer cleanup()

	createFile(tmpTestDir, "existing_file", "old content")

	type args struct {
		body   string
		suburl string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
	}{
		{"create new file", args{"New File Content", "new_file"}, 200, "New File Content"},
		{"update existing file", args{"updated content", "existing_file"}, 200, "updated content"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageBackend = Backend{tmpTestDir}
			router := chi.NewRouter()
			chi.RegisterMethod("UNLOCK")
			router.MethodFunc("UNLOCK", "/{id}", updateTfstate)
			ts := httptest.NewServer(router)
			defer ts.Close()

			rr, got := testRequest(t, ts, "UNLOCK", "/"+tt.args.suburl, strings.NewReader(tt.args.body))
			if rr.StatusCode != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.StatusCode, tt.wantStatus)
			}
			if got != tt.wantBody {
				t.Errorf("handler returned unexpected body: got %v want %v", got, tt.wantBody)
			}
		})
	}
}
