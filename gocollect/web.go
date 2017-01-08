// Copyright 2017 Landonia Ltd. All rights reserved.

package gocollect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// MaxRequestLength specifies the amount of bytes accepted on a request
// Do not let someone hang the service by sending continous stream of data
// on the request. 10Kb will be big enough for this example.
const MaxRequestLength int64 = 1024

// HandleHTTP will initialise the web interface
func HandleHTTP(addr string, store *Store) {

	// Create a gorilla server multiplexer
	r := mux.NewRouter()

	// Add a backup DB handler that makes it easy to create backups over HTTP
	r.HandleFunc("/backup", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="backup.db"`)
		w.Header().Set("Content-Length", strconv.FormatInt(store.Size(), 10))
		store.Backup(w)
	}).Methods(http.MethodGet)

	// Add in the handler to create a new user - we are only accepting json
	r.HandleFunc("/users", newHandler(store, addUser)).Methods(http.MethodPost).HeadersRegexp("Content-Type", "application/json")

	// Add in a search handler that allows the users to be searched
	r.HandleFunc("/users/search", newHandler(store, search)).Methods(http.MethodGet)

  // Add in a fuzzysearch handler that allows the users to be searched using a part of the email
	r.HandleFunc("/users/fuzzysearch", newHandler(store, fuzzysearch)).Methods(http.MethodGet)

	// Add the new handler that will return the user
	r.HandleFunc("/users/{userId}", newHandler(store, getUser)).Methods(http.MethodGet)

	// Add the handler to return the user events
	r.HandleFunc("/users/{userid}/events", newHandler(store, getUserEvent)).Methods(http.MethodGet)

	// Add the handler to add in new events
	r.HandleFunc("/users/{userid}/events", newHandler(store, addUserEvent)).Methods(http.MethodPost).HeadersRegexp("Content-Type", "application/json")

	// Setup the webserver
	s := &http.Server{
		Addr:         addr,
		Handler:      r,
		WriteTimeout: 1 * time.Minute,
		ReadTimeout:  1 * time.Minute,
	}

	// Attempt to start the service
	logger.Info("Starting gocollect server at address: %s", addr)

	// Start the server
	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Error("Server not shutdown gracefully: %s", err.Error())
		}
	}()
}

// storeHandler that will wrap the store for each handler
type storeHandler func(store *Store, w http.ResponseWriter, req *http.Request)

// newHandler will return a wrapper for the handler provided
func newHandler(store *Store, handler storeHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler(store, w, req)
	})
}

// addUser will handle any users being added to the store
func addUser(store *Store, w http.ResponseWriter, req *http.Request) {

	// Attempt to parse the body into a user
	var user User
	err := json.NewDecoder(io.LimitReader(req.Body, MaxRequestLength)).Decode(&user)
	if err != nil {
		logger.Error("Could not parse user data: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Attempt to add the user to the store
	id, err := store.AddUser(user)
	if err != nil {
		logger.Error("Could not add user data to store: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	logger.Info("Successfully added user data to store using id: %d", id)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("{\"id\":%d}", id)))
}

// getUser will attempt to return the user struct with a particular identifier
func getUser(store *Store, w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	rawID := vars["userId"]

	userID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		logger.Error("Could not parse userID: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	// Attempt to get the user with the ID
	var user User
	user, err = store.GetUser(userID)
	if err != nil {
		logger.Error("Could not find user with ID: %d: error: %s", userID, err.Error())
		w.WriteHeader(http.StatusNotFound)
	}

	// Write the user to the stream
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// search will search the fields for any matching records
// A query string param can be sent for each field to search
func search(store *Store, w http.ResponseWriter, req *http.Request) {

  // Get the email param
  email := req.URL.Query().Get("email")
  id, err := store.GetUserIDUsingEmail(email)
  if err != nil {
  		w.WriteHeader(http.StatusNotFound)
      return
  }
  // Write the id to the stream
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"id\":%d}", id)))
}

// SearchResult will be returned for any matching record IDs
type SearchResult struct {
  IDs []uint64 `json:"ids"`
}

// fuzzysearch will search the fields for any matching records based on the bytes
// that have been passed. So for example - "john.doe@"
func fuzzysearch(store *Store, w http.ResponseWriter, req *http.Request) {

  // Get the email param
  email := req.URL.Query().Get("email")
  ids, err := store.GetUserIDsMatchingFuzzyEmail(email)
  if err != nil {
  		w.WriteHeader(http.StatusNotFound)
      return
  }

  // Write the id to the stream
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
  if err = json.NewEncoder(w).Encode(SearchResult{ids}); err != nil {
    logger.Error("Error formatting data: %s", err.Error())
  }
}

func addUserEvent(store *Store, w http.ResponseWriter, req *http.Request) {
  // TODO
}

func getUserEvent(store *Store, w http.ResponseWriter, req *http.Request) {
  // TODO
}
