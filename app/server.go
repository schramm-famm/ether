package main

import (
	"ether/filesystem"
	"ether/handlers"
	"ether/kafka"
	"ether/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func logging(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f.ServeHTTP(w, r)
	})
}

func main() {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s)/?interpolateParams=true",
		os.Getenv("ETHER_DB_USERNAME"),
		os.Getenv("ETHER_DB_PASSWORD"),
		os.Getenv("ETHER_DB_LOCATION"))
	db, err := models.NewDB(connectionString)
	if err != nil {
		log.Fatal("Failed to open database: ", err)
		return
	}

	directory := filesystem.NewDirectory(os.Getenv("ETHER_CONTENT_DIR"))
	env := &handlers.Env{
		DB:        db,
		Directory: directory,
	}

	kafkaReader := kafka.NewReader(
		os.Getenv("ETHER_KAFKA_SERVER"),
		os.Getenv("ETHER_KAFKA_TOPIC"),
		directory,
		db,
	)
	go kafkaReader.Run()

	httpMux := mux.NewRouter()

	// Conversation CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations",
		env.PostConversationHandler,
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations",
		env.GetConversationsHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		env.GetConversationHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		env.PatchConversationHandler,
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		env.DeleteConversationHandler,
	).Methods("DELETE")

	// Conversation Content read
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/content",
		env.GetContentHandler,
	).Methods("GET")

	// User-Conversation Mapping CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		env.PostMappingHandler,
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		env.GetMappingHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		env.GetMappingsHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		env.PatchMappingHandler,
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		env.DeleteMappingHandler,
	).Methods("DELETE")
	httpMux.Use(logging)

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
