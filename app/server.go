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
		"%s:%s@tcp(%s)/%s?interpolateParams=true",
		os.Getenv("ETHER_DB_USERNAME"),
		os.Getenv("ETHER_DB_PASSWORD"),
		os.Getenv("ETHER_DB_LOCATION"),
		os.Getenv("ETHER_DB_DATABASE"))
	db, err := models.NewDB(connectionString)
	if err != nil {
		log.Fatal(err)
		return
	}

	env := &handlers.Env{db}

	httpMux := mux.NewRouter()

	// Conversation CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations",
		logging(env.PostConversationHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		logging(env.GetConversationHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		logging(env.PatchConversationHandler),
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		logging(env.DeleteConversationHandler),
	).Methods("DELETE")

	// User-Conversation Mapping CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		logging(env.PostMappingHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		logging(env.GetMappingHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		logging(env.GetMappingsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		logging(env.PatchMappingHandler),
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		logging(env.DeleteMappingHandler),
	).Methods("DELETE")

	httpSrv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
