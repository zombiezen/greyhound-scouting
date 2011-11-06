package main

import (
	htmltemplate "exp/template/html"
	"fmt"
	"http"
	"log"
	"os"
	"template"

	"launchpad.net/mgo"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

type Server struct {
	*mux.Router
	database  mgo.Database
	templates *template.Set
	Debug     bool
	Static    http.FileSystem
}

func NewServer(db mgo.Database) *Server {
	server := &Server{
		Router:    new(mux.Router),
		database:  db,
		templates: new(template.Set),
		Debug:     true,
	}
	server.templates.Funcs(template.FuncMap{
		"route": server.routeFunc(),
		"cycle": func(i int, vals ...interface{}) interface{} {
			return vals[i % len(vals)]
		},
	})
	return server
}

func (server *Server) TemplateSet() *template.Set {
	return server.templates
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

func (server *Server) routeFunc() func(string, ...string) (htmltemplate.URL, os.Error) {
	return func(name string, pairs ...string) (htmltemplate.URL, os.Error) {
		route, ok := server.NamedRoutes[name]
		if !ok {
			return "", fmt.Errorf("Could not resolve route %q", name)
		}
		url := route.URL(pairs...)
		if url == nil {
			return "", fmt.Errorf("Bad set of pairs for route %q: %v", name, pairs)
		}
		return htmltemplate.URL(url.String()), nil
	}
}

type Logger struct {
	http.Handler
}

func (logger Logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method, path := req.Method, req.URL.Path
	rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	logger.Handler.ServeHTTP(rec, req)
	log.Printf("%s %s %d %d", method, path, rec.statusCode, rec.size)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int64
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
