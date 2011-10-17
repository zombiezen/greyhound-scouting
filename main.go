package main

import (
	"flag"
	"http"
	"log"
	"os"

	"launchpad.net/mgo"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

func main() {
	mongoURL := flag.String("mongo", "localhost", "The URL for the MongoDB instance")
	database := flag.String("database", "scouting", "The database name in the MongoDB instance to use")
	address := flag.String("address", ":8080", "The address to listen for connections")
	flag.Parse()

	session, err := mgo.Mongo(*mongoURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	server := &Server{
		database: session.DB(*database),
		Debug:    true,
	}

	mux.Handle("/", server.Handler(hello))

	log.Printf("Listening on %s", *address)
	http.ListenAndServe(*address, Logger{mux.DefaultRouter})
}

func hello(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	w.Write([]byte("Hello, World!\n"))
	return nil
}

type Logger struct {
	http.Handler
}

func (logger Logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	logger.Handler.ServeHTTP(rec, req)
	log.Printf("%s %s %d %d", req.Method, req.URL.Path, rec.statusCode, rec.size)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size int64
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.ResponseWriter.WriteHeader(statusCode)
	rec.statusCode = statusCode
}

func (rec *responseRecorder) Write(p []byte) (n int, err os.Error) {
	n, err = rec.ResponseWriter.Write(p)
	rec.size += int64(n)
	return
}

type Server struct {
	database mgo.Database
	Debug    bool
}

func (server *Server) Session() *mgo.Session {
	return server.database.Session
}

func (server *Server) DB() mgo.Database {
	return server.database
}

func (server *Server) Handler(f func(*Server, http.ResponseWriter, *http.Request) os.Error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// TODO: This can be prettier.
		err := f(server, w, req)
		if err != nil {
			log.Printf("ERROR %s %s: %v", req.Method, req.URL.Path, err)
			if server.Debug {
				http.Error(w, err.String(), 500)
			} else {
				http.Error(w, "Server error encountered", 500)
			}
		}
	})
}
