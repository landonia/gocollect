// Copyright 2017 Landonia Ltd. All rights reserved.

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/landonia/gocollect/gocollect"
	"github.com/landonia/golog"
)

var (
	logger = golog.New("gollect.Main")

	// The application level params
	dbPath   = flag.String("db", "/usr/local/gocollect/bolt.db", "The path to the bolt DB")
	addr     = flag.String("addr", ":8080", "The address of the webserver")
	loglevel = flag.String("loglevel", "info", "The log level to use (off|fatal||error|warn|info|debug|trace)")
)

// bootstrap the application
func main() {
	flag.Parse()
	golog.LogLevel(*loglevel)

	// Create a new Store
	store := new(gocollect.Store)
	if err := store.Open(*dbPath); err != nil {
		logger.Fatal("Could not open the DB: %s", err.Error())
	}
	defer store.Close()

	// Initialise the store
	if err := store.Init(); err != nil {
		logger.Fatal("Could not initialise the store: %s", err)
	}

	// Block until we receive the exit
	// Wait for a shutdown signal
	exit := make(chan bool)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			logger.Info("Received exit signal - shutting down")
			exit <- true
		}()
	}()

	// Start the web interface
	gocollect.HandleHTTP(*addr, store)

	// Wait to shutdown
	<-exit
	logger.Info("Shutdown gollect service at address: %s", *addr)
}
