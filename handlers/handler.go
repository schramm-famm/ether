package handlers

import (
	"ether/models"
	"log"
	"net/http"
)

// Env represents all application-level items that are needed by handlers
type Env struct {
	DB models.Datastore
}

func internalServerError(w http.ResponseWriter, err error) {
	errMsg := "Internal Server Error"
	log.Println(errMsg + ": " + err.Error())
	http.Error(w, errMsg, http.StatusInternalServerError)
}
