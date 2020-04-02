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

	client := &http.Client{}
	karen := os.Getenv("KAREN_SERVER")

	kafkaEnv := &kafka.Env{
		DB:           db,
		CachedWriter: filesystem.NewCachedWriter(directory),
	}

	// Start file writer goroutine
	go kafkaEnv.CachedWriter.Run()

	kafkaReader := kafka.NewReader(
		os.Getenv("ETHER_KAFKA_SERVER"),
		os.Getenv("ETHER_KAFKA_TOPIC"),
	)

	// Start Kafka reader goroutine
	go kafkaReader.Run(kafkaEnv.ProcessWSMessage)

	httpEnv := &handlers.Env{
		DB:        db,
		Directory: directory,
		Client:    client,
		KarenHost: karen,
	}

	httpMux := mux.NewRouter()

	// Conversation CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations",
		httpEnv.PostConversationHandler,
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations",
		httpEnv.GetConversationsHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		httpEnv.GetConversationHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		httpEnv.PatchConversationHandler,
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}",
		httpEnv.DeleteConversationHandler,
	).Methods("DELETE")

	// Conversation Content read
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/content",
		httpEnv.GetContentHandler,
	).Methods("GET")

	// User-Conversation Mapping CRUD
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		httpEnv.PostMappingHandler,
	).Methods("POST")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		httpEnv.GetMappingHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users",
		httpEnv.GetMappingsHandler,
	).Methods("GET")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		httpEnv.PatchMappingHandler,
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/ether/v1/conversations/{conversation_id:[0-9]+}/users/{user_id:[0-9]+}",
		httpEnv.DeleteMappingHandler,
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
