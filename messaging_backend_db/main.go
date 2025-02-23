// Package main initializes the server for the database.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/authorize"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/handler"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// main sets up the server, database, schema and initialized the server to the specified port.
func main() {
	//varaibles for server
	var server http.Server
	var port int // used as the port num
	var err error

	// varaibles for flags
	var docSchema string
	var tokenFile string

	//defining flags
	// Specify the port your server should listen on with defualt value as 3318
	flag.IntVar(&port, "p", 3318, "port number")
	flag.StringVar(&docSchema, "s", "error", "JSON schema file name")
	flag.StringVar(&tokenFile, "t", "", "file name for token")

	flag.Parse()

	// Create a JSON Schema compiler
	compiler := jsonschema.NewCompiler()

	// Compile JSON schema
	schema, err := compiler.Compile(docSchema)
	if err != nil {
		slog.Error("schema compilation error", "error", err)
		return
	}

	// initialize the owlDB database and token map
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	authorize.Initialize(tokenFile, tokenMap)

	// The following code should go last and remain unchanged.
	// Note that you must actually initialize 'server' and 'port'
	// before this.

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			slog.Info("handler function is called")
			handler.HndlRequest(w, r, &owlDB, tokenMap, schema)
		})

	server = http.Server{
		Addr:    ":" + fmt.Sprintf("%d", port),
		Handler: handler,
	}

	// signal.Notify requires the channel to be buffered
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt, syscall.SIGTERM)
	go func() {
		// Wait for Ctrl-C signal
		<-ctrlc
		server.Close()
	}()

	// Start server
	slog.Info("Listening", "port", port)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Server closed", "error", err)
	} else {
		slog.Info("Server closed", "error", err)
	}
}
