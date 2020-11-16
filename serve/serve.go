package serve

import (
	"fmt"
	"log"
	"net/http"
	"suggestions/datastore"

	"github.com/gorilla/mux"
)

type Server struct {
	Datastore  *datastore.Datastore
	Cachestore *datastore.Cachestore
}

var server *Server

func NewServer() *Server {
	ds, err := datastore.NewDatastore()
	if err != nil {
		log.Fatalf("couldn't create datastore: %v", err)
	}
	cs := datastore.NewCachestore()

	server := &Server{
		Datastore:  ds,
		Cachestore: cs,
	}
	return server
}

func homeHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func Serve(port string) error {
	server = NewServer()

	router := mux.NewRouter().StrictSlash(true)

	// Create the route
	router.HandleFunc("/", homeHandle)
	router.HandleFunc("/query/{query}", queryHandle)
	log.Fatal(http.ListenAndServe(port, router))
	return nil
}
