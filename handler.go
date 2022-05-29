package urlshort

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
)

var (
	bucketName     = []byte("urls")
	CreatePostPath = "/create"
)

// BoltHandler searches for the redirect routing of the passed short URL
// if it not finds one
func BoltHandler(db *bolt.DB, fallback http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		var shouldFallback bool
		var redirectUrl string
		/*
			Transactions should not be dependent on one another.
			Opening a read transaction and a write transaction in the same goroutine can cause the writer to deadlock
			because the database periodically needs to re-mmap itself as it grows and it cannot do that
			while a read transaction is open.
			The mistake was to fallback to an Update tx inside a View tx, you should always fallback outside of the View tx
		*/
		_ = db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(bucketName)
			url := bucket.Get([]byte(path))
			if url == nil {
				shouldFallback = true
				return nil
			}
			redirectUrl = string(url)
			return nil
		})
		if shouldFallback {
			fallback.ServeHTTP(writer, request)
			return
		}
		http.Redirect(writer, request, redirectUrl, http.StatusTemporaryRedirect)
	}
}

// CreateHandler writes new short URLs to the db
func CreateHandler(db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var pu pathUrl
		err := json.NewDecoder(r.Body).Decode(&pu)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if pu.Path == CreatePostPath {
			http.Error(w, "Path prohibited to use", http.StatusBadRequest)
			return
		}

		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(bucketName)
			bucket.Put([]byte(pu.Path), []byte(pu.Url))
			return nil
			//return nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write([]byte("Register updated. Short URL: " + pu.Path))
		return
	}
}

type pathUrl struct {
	Path string
	Url  string
}

// InitDBwYaml writes to the db all the routing elements from yaml
func InitDBwYaml(db *bolt.DB, path *string) {
	fc, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalf("Some error occured while trying to read this file: %s\n%v", *path, err)
	}
	yamlUrls, err := parseYml(fc)
	if err != nil {
		log.Fatalf("Something went wrong while parsing the yaml: %v\n", err)
	}
	initializeDb(yamlUrls, db)
}

func parseYml(yml []byte) (yamlUrls []pathUrl, err error) {
	err = yaml.Unmarshal(yml, &yamlUrls)
	if err != nil {
		return nil, err
	}
	return
}

// initializeDb creates some initial values
func initializeDb(pathsToUrls []pathUrl, db *bolt.DB) {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		for _, e := range pathsToUrls {
			if err := bucket.Put([]byte(e.Path), []byte(e.Url)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
