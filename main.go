package main

import (
	"code.google.com/p/gorilla/mux"
	"encoding/csv"
	"flag"
	"io"
	"launchpad.net/mgo"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const templatePrefix = "templates/"

var server *Server

// Flags
var (
	mongoURL  string
	database  string
	address   string
	imagedir  string
	staticdir string
	debug     bool
)

func main() {
	parseFlags()

	if flag.NArg() == 0 {
		createServer()
		parseTemplates()
		addRoutes()

		log.Printf("Listening on %s", address)
		http.ListenAndServe(address, Logger{server})
	} else {
		switch flag.Arg(0) {
		case "teams":
			importTeams()
		case "schedule":
			importSchedule()
		default:
			log.Fatal("usage: scouting [teams]")
		}
	}
}

func parseFlags() {
	flag.StringVar(&mongoURL, "mongo", "localhost", "The URL for the MongoDB instance")
	flag.StringVar(&database, "database", "scouting", "The database name in the MongoDB instance to use")
	flag.StringVar(&address, "address", ":8080", "The address to listen for connections")
	flag.StringVar(&staticdir, "staticdir", "static", "The directory to serve static files from")
	flag.StringVar(&imagedir, "imagedir", "images", "The directory to serve team images from")
	flag.BoolVar(&debug, "debug", false, "Display extra information in-browser about the program")
	flag.Parse()
}

func openDatastore() (Datastore, error) {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return nil, err
	}
	return mongoDatastore{session.DB(database)}, nil
}

func createServer() {
	datastore, err := openDatastore()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	server = NewServer(datastore)
	server.imagestore = directoryImagestore{imagedir, &url.URL{Path: "/team/images/"}}
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

	server.Handle("/static{path:/.*}", makeStaticHandler(http.Dir(staticdir))).Name("static")
	server.Handle("/team/images{path:/.*}", makeStaticHandler(http.Dir(imagedir))).Name("teamImages")
}

// makeStaticHandler returns a new HTTP handler that uses the Gorilla router 'path' variable.
func makeStaticHandler(fs http.FileSystem) http.Handler {
	s := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		req.URL.Path = "/" + vars["path"]
		s.ServeHTTP(w, req)
	})
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

// importTeams handles the teams command.
func importTeams() {
	if flag.NArg() != 1 {
		log.Fatal("usage: scouting teams")
	}

	datastore, err := openDatastore()
	if err != nil {
		log.Fatalln("Could not connect to database:", err)
	}

	r := csv.NewReader(os.Stdin)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		if len(row) != 2 {
			log.Fatal("Team CSV files must be number,name")
		}

		num, err := strconv.Atoi(row[0])
		if err != nil {
			log.Printf("Row found with bad number: %q (%v)", row[0], err)
			continue
		}

		if err := datastore.UpsertTeam(&Team{Number: num, Name: row[1]}); err != nil {
			log.Printf("Error updating team: %v", err)
		}
	}
}

// importSchedule handles the schedule command.
func importSchedule() {
	var event *Event
	var etag EventTag
	var err error

	switch flag.NArg() {
	case 2:
		if etag, err = ParseEventTag(flag.Arg(1)); err != nil {
			log.Fatalf("Invalid code %q: %v", flag.Arg(1), err)
		}
	case 4:
		// Parse/validate date
		dateParts := strings.SplitN(flag.Arg(1), "-", 3)
		if len(dateParts) != 3 {
			log.Fatal("DATE must be YYYY-MM-DD, got %q", flag.Arg(1))
		}
		year, err := strconv.ParseUint(dateParts[0], 10, 0)
		if err != nil {
			log.Fatalf("Bad year: %v", err)
		}
		month, err := strconv.ParseUint(dateParts[1], 10, 0)
		if err != nil {
			log.Fatalf("Bad month: %v", err)
		}
		day, err := strconv.ParseUint(dateParts[2], 10, 0)
		if err != nil {
			log.Fatalf("Bad day: %v", err)
		}
		d := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		year, month, day = uint64(d.Year()), uint64(d.Month()), uint64(d.Day())

		// Validate location code
		if flag.Arg(2) == "" {
			log.Fatal("Location code must be non-empty")
		}

		// Upsert event
		etag = EventTag{flag.Arg(2), uint(year)}
		event = new(Event)
		event.Location.Code = flag.Arg(2)
		event.Location.Name = flag.Arg(3)
		event.Date.Year = int(year)
		event.Date.Month = int(month)
		event.Date.Day = int(day)
	default:
		log.Fatal("usage: scouting schedule ( CODE | DATE LOC_CODE LOC_NAME )")
	}

	// Open datastore
	datastore, err := openDatastore()
	if err != nil {
		log.Fatalln("Could not connect to database:", err)
	}

	// Create event
	teamSet := make(map[int]bool)
	if flag.NArg() == 2 {
		event, err = datastore.FetchEvent(etag)
		if err != nil {
			log.Fatalf("Fetching event %q: %v", flag.Arg(1), err)
		}
		for _, num := range event.Teams {
			teamSet[num] = true
		}
	}

	// Read matches
	r := csv.NewReader(os.Stdin)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("CSV file error: %v (stopping match import, creating event)")
			break
		}

		if len(row) != 9 {
			log.Fatal("Bad CSV file: must be time,type,num,red1,red2,red3,blue1,blue2,blue3")
		}

		nums := make([]uint, 7)
		matchType := MatchType(row[1])
		if matchType != Qualification && matchType != QuarterFinal && matchType != SemiFinal && matchType != Final {
			log.Printf("Bad match type %q: must be %q/%q/%q/%q", matchType, Qualification, QuarterFinal, SemiFinal, Final)
			continue
		}
		for i, col := range row[2:] {
			n, err := strconv.ParseUint(col, 10, 0)
			if err != nil {
				log.Printf("Bad cell %q: %v", col, err)
			}
			nums[i] = uint(n)
		}
		m := Match{Type: matchType, Number: int(nums[0])}
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[1]), Alliance: Red})
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[2]), Alliance: Red})
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[3]), Alliance: Red})
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[4]), Alliance: Blue})
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[5]), Alliance: Blue})
		m.Teams = append(m.Teams, TeamInfo{Team: int(nums[6]), Alliance: Blue})
		if err = datastore.UpsertMatch(etag, &m); err != nil {
			log.Printf("Error for match %s %d: %v", m.Type, m.Number, err)
			continue
		}

		for _, n := range nums[1:] {
			teamSet[int(n)] = true
		}
	}

	// Update event
	event.Teams = make([]int, 0, len(teamSet))
	for teamNum, _ := range teamSet {
		event.Teams = append(event.Teams, teamNum)
	}
	sort.Ints(event.Teams)
	if err := datastore.UpsertEvent(event); err != nil {
		log.Fatal("Upserting event: %v", err)
	}
}
