package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f(w, r)
	}
}

func main() {
	httpMux := mux.NewRouter()

	httpSrv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
