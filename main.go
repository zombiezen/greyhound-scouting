package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"launchpad.net/mgo"

	"code.google.com/p/gorilla/gorilla/mux"
)

const templatePrefix = "templates/"

const eventURLPrefix = "/event/{year:[1-9][0-9]*}/{location:[a-z]+}/"
const matchURLPrefix = eventURLPrefix + "match/{matchType:qualification|quarter|semifinal|final}/{matchNumber:[1-9][0-9]*}/"

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

	if _, err := server.Templates().ParseGlob(templatePrefix + "*.html"); err != nil {
		log.Fatalf("Could not load templates: %v", err)
	}
	if _, err := server.Templates().ParseFiles(templatePrefix + "gopher"); err != nil {
		log.Fatalf("Could not load gopher: %v", err)
	}

	server.Handle("/", server.Handler(index)).Name("root")
	server.Handle("/jump", server.Handler(jump)).Name("jump")

	server.Handle("/team/", server.Handler(teamIndex)).Name("team.index").RedirectSlash(true)
	server.Handle("/team/{number:[1-9][0-9]*}/", server.Handler(viewTeam)).Name("team.view").RedirectSlash(true)

	server.Handle("/event/", server.Handler(eventIndex)).Name("event.index").RedirectSlash(true)
	server.Handle(eventURLPrefix, server.Handler(viewEvent)).Name("event.view").RedirectSlash(true)
	server.Handle(eventURLPrefix+"scout-forms.pdf", server.Handler(eventScoutForms)).Name("event.scoutForms")

	server.Handle(matchURLPrefix, server.Handler(viewMatch)).Name("match.view").RedirectSlash(true)

	staticServer := http.FileServer(http.Dir(*staticdir))
	server.HandleFunc("/static{path:/.*}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		req.URL.Path = "/" + vars["path"]
		staticServer.ServeHTTP(w, req)
	}).Name("static")

	log.Printf("Listening on %s", *address)
	http.ListenAndServe(*address, Logger{server})
}

func index(server *Server, w http.ResponseWriter, req *http.Request) error {
	return server.Templates().ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
	})
}

func jump(server *Server, w http.ResponseWriter, req *http.Request) error {
	query := req.FormValue("q")
	log.Printf("Jump %q", query)
	if query != "" {
		if _, err := strconv.Atoi(query); err == nil {
			// Team number
			http.Redirect(w, req, server.NamedRoutes["team.view"].URL("number", query).String(), http.StatusFound)
			return nil
		}

		if eventTag, err := ParseEventTag(query); err == nil {
			// Event
			http.Redirect(w, req, server.NamedRoutes["event.view"].URL(
				"year", strconv.FormatUint(uint64(eventTag.Year), 10),
				"location", eventTag.LocationCode,
			).String(), http.StatusFound)
			return nil
		}

		// TODO: other tags
	}
	return server.Templates().ExecuteTemplate(w, "jump.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Query":   query,
	})
}
