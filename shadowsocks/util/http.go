package util

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

var r *mux.Router

func SetRoute(r *mux.Router) {
	r = r
}

func HTTPError(w http.ResponseWriter, desc string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, desc)
}

func HTTPSuccess(w http.ResponseWriter, desc string) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, desc)
}
