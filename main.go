package main

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

var storageBackend Backend
var config Config

func getTfstate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var e *fs.PathError

	tfID := chi.URLParam(r, "id")
	if body, err = storageBackend.get(tfID); err != nil {
		if errors.As(err, &e) {
			var operation = err.(*fs.PathError).Op
			if operation == "stat" || operation == "CreateFile" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(http.StatusText(http.StatusNotFound)))
			} else {
				logger.Warnf("Can not Access File: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			}
			return
		}
		logger.Warnf("Complete unexpected Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func updateTfstate(w http.ResponseWriter, r *http.Request) {
	tfID := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if err := storageBackend.update(tfID, reqBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(reqBody)
}

func purgeTfstate(w http.ResponseWriter, r *http.Request) {
	tfID := chi.URLParam(r, "id")
	if err := storageBackend.purge(tfID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	_, _ = w.Write([]byte("{\"state\": \"tfstate deleted\"}"))
}

func lockTfstate(w http.ResponseWriter, r *http.Request) {
	var lockFile []byte
	var err error
	var conflict *ConflictError

	tfID := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if lockFile, err = storageBackend.lock(tfID, reqBody); err != nil {
		if errors.As(err, &conflict) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	_, _ = w.Write(lockFile)
}

func unlockTfstate(w http.ResponseWriter, r *http.Request) {
	var conflict *ConflictError

	tfID := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if err := storageBackend.unlock(tfID, reqBody); err != nil {
		if errors.As(err, &conflict) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	_, _ = w.Write(reqBody)
}

func handleRequests() {
	logger.Debugf("current storage path: %s", config.storageDirectory)
	storageBackend = Backend{dir: config.storageDirectory}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if config.authEnabled {
		r.Use(middleware.BasicAuth("restricted access", config.getAuthMap()))
	}

	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Get("/{id}", getTfstate)
	r.Post("/{id}", updateTfstate)
	r.Delete("/{id}", purgeTfstate)
	r.MethodFunc("LOCK", "/{id}", lockTfstate)
	r.MethodFunc("UNLOCK", "/{id}", unlockTfstate)
	logger.Fatal(http.ListenAndServe(config.getAddr(), r))
}

func init() {
	SetLogger(logrus.New())
	config.loadConfig(".env")
}

func main() {
	handleRequests()
}
