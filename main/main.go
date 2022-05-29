package main

import (
	"flag"
	"fmt"
	"github.com/Fillip-Molodtsov-gophercising/urlshort"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	serverAddress  = ":8080"
	fileFlag       = "file"
	dbName         = "url.db"
	ymlFileDefault = "example.yml"
)

func main() {
	path := flag.String(fileFlag, defaultYml(), "The file where default short URLs are stored")
	flag.Parse()
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	urlshort.InitDBwYaml(db, path)

	createHandler := urlshort.CreateHandler(db)
	mux := appInitialMux()
	mux.HandleFunc(urlshort.CreatePostPath, createHandler)
	boltHandler := urlshort.BoltHandler(db, mux)

	fmt.Printf("Starting the server on %s\n", serverAddress)
	http.ListenAndServe(serverAddress, boltHandler)
}

func appInitialMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func defaultYml() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot find working directory: %v", err)
	}
	return filepath.Join(pwd, ymlFileDefault)
}
