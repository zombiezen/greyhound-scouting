package main

import (
	htmltemplate "exp/template/html"
	"flag"
	"http"
	"log"
	"os"
	"strconv"

	"launchpad.net/mgo"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

const templatePrefix = "templates/"

var escapedTemplates = []string{
	"index.html",
	"jump.html",
	"team.html",
	"team-index.html",
	"event.html",
	"event-index.html",
	"error.html",
	"error-debug.html",
}

func main() {
	mongoURL := flag.String("mongo", "localhost", "The URL for the MongoDB instance")
	database := flag.String("database", "scouting", "The database name in the MongoDB instance to use")
	address := flag.String("address", ":8080", "The address to listen for connections")
	staticdir := flag.String("staticdir", "static", "The directory to serve static files from")
	debug := flag.Bool("debug", false, "Display extra information in-browser about the program")
	flag.Parse()

	session, err := mgo.Mongo(*mongoURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	server := NewServer(mongoDatastore{session.DB(*database)})
	server.Debug = *debug

	if _, err := server.TemplateSet().ParseGlob(templatePrefix + "sets/*.html"); err != nil {
		log.Fatalf("Could not load template sets: %v", err)
	}
	if _, err := server.TemplateSet().ParseTemplateGlob(templatePrefix + "*.html"); err != nil {
		log.Fatalf("Could not load templates: %v", err)
	}
	if _, err := server.TemplateSet().ParseTemplateFiles(templatePrefix + "gopher"); err != nil {
		log.Fatalf("Could not load gopher: %v", err)
	}
	if _, err := htmltemplate.EscapeSet(server.TemplateSet(), escapedTemplates...); err != nil {
		log.Fatalf("Could not autoescape templates: %v", err)
	}

	server.Handle("/", server.Handler(index)).Name("root")
	server.Handle("/jump", server.Handler(jump)).Name("jump")

	server.Handle("/team/", server.Handler(teamIndex)).Name("team.index").RedirectSlash(true)
	server.Handle("/team/{number:[1-9][0-9]*}/", server.Handler(viewTeam)).Name("team.view").RedirectSlash(true)

	server.Handle("/event/", server.Handler(eventIndex)).Name("event.index").RedirectSlash(true)
	server.Handle("/event/{year:[1-9][0-9]*}/{location:[a-z]+}/", server.Handler(viewEvent)).Name("event.view").RedirectSlash(true)

	staticServer := http.FileServer(http.Dir(*staticdir))
	server.HandleFunc("/static{path:/.*}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		req.URL.Path = "/" + vars["path"]
		staticServer.ServeHTTP(w, req)
	}).Name("static")

	log.Printf("Listening on %s", *address)
	http.ListenAndServe(*address, Logger{server})
}

func index(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	return server.TemplateSet().Execute(w, "index.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
	})
}

func jump(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	query := req.FormValue("q")
	log.Printf("Jump %q", query)
	if query != "" {
		if _, err := strconv.Atoi(query); err == nil {
			// Team number
			http.Redirect(w, req, server.NamedRoutes["team.view"].URL("number", query).String(), http.StatusFound)
			return nil
		}

		// TODO: other tags
	}
	return server.TemplateSet().Execute(w, "jump.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Query":   query,
	})
}
