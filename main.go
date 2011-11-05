package main

import (
	htmltemplate "exp/template/html"
	"flag"
	"http"
	"log"
	"os"

	"launchpad.net/mgo"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

const templatePrefix = "templates/"

var escapedTemplates = []string{
	"index.html",
}

func main() {
	mongoURL := flag.String("mongo", "localhost", "The URL for the MongoDB instance")
	database := flag.String("database", "scouting", "The database name in the MongoDB instance to use")
	address := flag.String("address", ":8080", "The address to listen for connections")
	staticdir := flag.String("staticdir", "static", "The directory to serve static files from")
	flag.Parse()

	session, err := mgo.Mongo(*mongoURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	server := NewServer(session.DB(*database))
	server.Static = http.Dir(*staticdir)

	if _, err := server.TemplateSet().ParseGlob(templatePrefix + "sets/*.html"); err != nil {
		log.Fatalf("Could not load template sets: %v", err)
	}
	if _, err := server.TemplateSet().ParseTemplateGlob(templatePrefix + "*.html"); err != nil {
		log.Fatalf("Could not load templates: %v", err)
	}
	if _, err := htmltemplate.EscapeSet(server.TemplateSet(), escapedTemplates...); err != nil {
		log.Fatalf("Could not autoescape templates: %v", err)
	}

	server.Handle("/", server.Handler(index)).Name("root")
	server.Handle("/jump", server.Handler(index)).Name("jump")
	server.Handle("/static{path:/.*}", server.Handler(staticFile)).Name("static")

	log.Printf("Listening on %s", *address)
	http.ListenAndServe(*address, Logger{server})
}

func index(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	server.TemplateSet().Execute(w, "index.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
	})
	return nil
}

func staticFile(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)
	// TODO: Don't do an allocation every time.
	fs := http.FileServer(server.Static)
	req.URL.Path = "/" + vars["path"]
	fs.ServeHTTP(w, req)
	return nil
}
