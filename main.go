package main

import (
	"code.google.com/p/gorilla/mux"
	"flag"
	"launchpad.net/mgo"
	"log"
	"net/http"
	"strconv"
)

const templatePrefix = "templates/"

var server *Server

// Flags
var (
	mongoURL  string
	database  string
	address   string
	staticdir string
	debug     bool
)

func main() {
	parseFlags()
	createServer()
	parseTemplates()
	addRoutes()

	log.Printf("Listening on %s", address)
	http.ListenAndServe(address, Logger{server})
}

func parseFlags() {
	flag.StringVar(&mongoURL, "mongo", "localhost", "The URL for the MongoDB instance")
	flag.StringVar(&database, "database", "scouting", "The database name in the MongoDB instance to use")
	flag.StringVar(&address, "address", ":8080", "The address to listen for connections")
	flag.StringVar(&staticdir, "staticdir", "static", "The directory to serve static files from")
	flag.BoolVar(&debug, "debug", false, "Display extra information in-browser about the program")
	flag.Parse()
}

func createServer() {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	server = NewServer(mongoDatastore{session.DB(database)})
	server.Debug = debug
}

func parseTemplates() {
	if _, err := server.Templates().ParseGlob(templatePrefix + "*.html"); err != nil {
		log.Fatalf("Could not load templates: %v", err)
	}
	if _, err := server.Templates().ParseFiles(templatePrefix + "gopher"); err != nil {
		log.Fatalf("Could not load gopher: %v", err)
	}
}

func addRoutes() {
	server.Handle("/", server.Handler(index)).Name("root")
	server.Handle("/jump", server.Handler(jump)).Name("jump")

	teamRouter := server.PathPrefix("/team").Subrouter()
	teamRouter.Handle("/", server.Handler(teamIndex)).Name("team.index")
	teamRouter.Handle("/{number:[1-9][0-9]*}/", server.Handler(viewTeam)).Name("team.view")

	eventRootRouter := server.PathPrefix("/event").Subrouter()
	eventRootRouter.Handle("/", server.Handler(eventIndex)).Name("event.index")

	eventRouter := eventRootRouter.PathPrefix("/{year:[1-9][0-9]*}/{location:[a-z]+}").Subrouter()
	eventRouter.Handle("/", server.Handler(viewEvent)).Name("event.view")
	eventRouter.Handle("/scout-forms.pdf", server.Handler(eventScoutForms)).Name("event.scoutForms")
	eventRouter.Handle("/teams.csv", server.Handler(eventSpreadsheet)).Name("event.spreadsheet")

	matchRouter := eventRouter.PathPrefix("/match/{matchType:qualification|quarter|semifinal|final}/{matchNumber:[1-9][0-9]*}").Subrouter()
	matchRouter.Handle("/", server.Handler(viewMatch)).Name("match.view")
	matchRouter.Handle("/match-sheet.pdf", server.Handler(matchSheet)).Name("match.sheet")
	matchRouter.Handle("/+edit/{teamNumber:[1-9][0-9]*}", server.Handler(editMatchTeam)).Name("match.editTeam")

	staticServer := http.FileServer(http.Dir(staticdir))
	server.HandleFunc("/static{path:/.*}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		req.URL.Path = "/" + vars["path"]
		staticServer.ServeHTTP(w, req)
	}).Name("static")
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
			u, err := server.GetRoute("team.view").URL("number", query)
			if err != nil {
				return err
			}
			http.Redirect(w, req, u.String(), http.StatusFound)
			return nil
		}

		if eventTag, err := ParseEventTag(query); err == nil {
			// Event
			u, err := server.GetRoute("event.view").URL(
				"year", strconv.FormatUint(uint64(eventTag.Year), 10),
				"location", eventTag.LocationCode,
			)
			if err != nil {
				return err
			}
			http.Redirect(w, req, u.String(), http.StatusFound)
			return nil
		}

		if matchTag, err := ParseMatchTag(query); err == nil {
			// Match
			u, err := server.GetRoute("match.view").URL(
				"year", strconv.FormatUint(uint64(matchTag.Year), 10),
				"location", matchTag.LocationCode,
				"matchType", string(matchTag.MatchType),
				"matchNumber", strconv.FormatUint(uint64(matchTag.MatchNumber), 10),
			)
			if err != nil {
				return err
			}
			http.Redirect(w, req, u.String(), http.StatusFound)
			return nil
		}

		if matchTeamTag, err := ParseMatchTeamTag(query); err == nil {
			// Edit Match Team
			u, err := server.GetRoute("match.editTeam").URL(
				"year", strconv.FormatUint(uint64(matchTeamTag.Year), 10),
				"location", matchTeamTag.LocationCode,
				"matchType", string(matchTeamTag.MatchType),
				"matchNumber", strconv.FormatUint(uint64(matchTeamTag.MatchNumber), 10),
				"teamNumber", strconv.FormatUint(uint64(matchTeamTag.TeamNumber), 10),
			)
			if err != nil {
				return err
			}
			http.Redirect(w, req, u.String(), http.StatusFound)
			return nil
		}
	}
	return server.Templates().ExecuteTemplate(w, "jump.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Query":   query,
	})
}
