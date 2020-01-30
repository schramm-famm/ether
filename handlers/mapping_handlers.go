package handlers

import "net/http"

// PostMappingHandler adds a single user to a conversation
func (env *Env) PostMappingHandler(w http.ResponseWriter, r *http.Request) {}

// GetMappingHandler gets a single user from a conversation
func (env *Env) GetMappingHandler(w http.ResponseWriter, r *http.Request) {}

// GetMappingsHandler gets all users from a conversation
func (env *Env) GetMappingsHandler(w http.ResponseWriter, r *http.Request) {}

// PatchMappingHandler updates a single user in a conversation
func (env *Env) PatchMappingHandler(w http.ResponseWriter, r *http.Request) {}

// DeleteMappingHandler deletes a single user from a conversation
func (env *Env) DeleteMappingHandler(w http.ResponseWriter, r *http.Request) {}
