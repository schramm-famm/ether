package main

import (
	"ether/handlers"
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
	httpMux.HandleFunc(
		"/api/conversations",
		logging(handlers.PostConversationsHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(handlers.GetConversationsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(handlers.PutConversationsHandler),
	).Methods("PUT")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(handlers.DeleteConversationsHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(handlers.PatchConversationsHandler),
	).Methods("PATCH")

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
