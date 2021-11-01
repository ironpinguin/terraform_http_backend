package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func createDirectory() (string, func()) {
	tempdir, _ := os.MkdirTemp(os.TempDir(), "testrun*")
	directory := tempdir + string(os.PathSeparator)

	return directory, func() {
		_ = os.RemoveAll(tempdir)
	}
}
func createFile(directory string, filename string, content string) {
	_ = os.WriteFile(directory+filename, []byte(content), 0644)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return resp, string(respBody)
}

func checkLogMessage(t *testing.T, wantLogs []string, hooks *test.Hook) {
	if wantLogs != nil {
		entries := hooks.AllEntries()
		for _, logMessage := range wantLogs {
			found := false
			var position int
			for id, entry := range entries {
				if !found {
					if logMessage == entry.Message {
						found = true
						position = id
					}
				}
			}
			assert.Contains(t, entries[position].Message, logMessage)
		}
	}
	hooks.Reset()
}
