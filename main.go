package main

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var storage_backend Backend

func get_tfstate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var e *fs.PathError

	tf_id := chi.URLParam(r, "id")
	if body, err = storage_backend.get(tf_id); err != nil {
		if errors.As(err, &e) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func update_tfstate(w http.ResponseWriter, r *http.Request) {
	tf_id := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if err := storage_backend.update(tf_id, reqBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(reqBody)
}

func purge_tfstate(w http.ResponseWriter, r *http.Request) {
	tf_id := chi.URLParam(r, "id")
	if err := storage_backend.pruge(tf_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}

func lock_tfstate(w http.ResponseWriter, r *http.Request) {
	var lock_file []byte
	var err error
	var conflict *ConflictError

	tf_id := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if lock_file, err = storage_backend.lock(tf_id, reqBody); err != nil {
		if errors.As(err, &conflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Write(lock_file)
}

func unlock_tfstate(w http.ResponseWriter, r *http.Request) {
	var conflict *ConflictError

	tf_id := chi.URLParam(r, "id")
	reqBody, _ := ioutil.ReadAll(r.Body)
	if err := storage_backend.unlock(tf_id, reqBody); err != nil {
		if errors.As(err, &conflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(http.StatusText(http.StatusConflict)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Write(reqBody)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{\"dfasjk\": \"djskalf\"}"))
}

func handleRequests() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Get("/", hello)
	r.Get("/{id}", get_tfstate)
	r.Post("/{id}", update_tfstate)
	r.Delete("/{id}", purge_tfstate)
	r.MethodFunc("LOCK", "/{id}", lock_tfstate)
	r.MethodFunc("UNLOCK", "/{id}", unlock_tfstate)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	pwd, _ := os.Getwd()
	storage_backend = Backend{pwd + string(os.PathSeparator) + "store" + string(os.PathSeparator)}
	log.Warn(pwd + string(os.PathSeparator) + "store")
	handleRequests()
}
