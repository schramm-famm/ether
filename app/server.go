package main

import (
	"ether/handlers"
	"ether/models"
	"fmt"
	"log"
	"net/http"
	"os"
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
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		os.Getenv("ETHER_DB_USERNAME"),
		os.Getenv("ETHER_DB_PASSWORD"),
		os.Getenv("ETHER_DB_LOCATION"),
		os.Getenv("ETHER_DB_DATABASE"))
	db, err := models.NewDB(connectionString)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("%v", db)

	env := &handlers.Env{db}

	httpMux := mux.NewRouter()
	httpMux.HandleFunc(
		"/api/conversations",
		logging(env.PostConversationsHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(env.GetConversationsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(env.DeleteConversationsHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/api/conversations",
		logging(env.PatchConversationsHandler),
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
