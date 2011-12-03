package main

import (
	"http"
	"os"
	"strconv"
	"time"

	"bitbucket.org/zombiezen/gopdf/pdf"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

func eventIndex(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	// Query for events
	events := server.Store().Events(int(time.LocalTime().Year))

	// Fetch events
	var eventList []Event
	if err := events.Limit(50).All(&eventList); err != nil {
		return err
	}

	// Render page
	return server.TemplateSet().Execute(w, "event-index.html", map[string]interface{}{
		"Server":    server,
		"Request":   req,
		"EventList": eventList,
	})
}

func routeEventTag(vars mux.RouteVars) EventTag {
	year, _ := strconv.Atoui(vars["year"])
	return EventTag{
		Year:         year,
		LocationCode: vars["location"],
	}
}

func viewEvent(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Fetch matches
	matches, err := server.Store().FetchMatches(event.Tag())
	if err != nil {
		return err
	}

	// Fetch teams
	teams, err := server.Store().FetchTeams(event.Teams)
	if err != nil {
		return err
	}

	return server.TemplateSet().Execute(w, "event.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Event":   event,
		"Matches": matches,
		"Teams":   teams,
	})
}

func routeMatchTag(vars mux.RouteVars) MatchTag {
	num, _ := strconv.Atoui(vars["matchNumber"])
	return MatchTag{
		EventTag:    routeEventTag(vars),
		MatchType:   MatchType(vars["matchType"]),
		MatchNumber: num,
	}
}

func viewMatch(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)

	// Fetch match
	match, err := server.Store().FetchMatch(routeMatchTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	return server.TemplateSet().Execute(w, "match.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Match": match,
	})
}

func eventScoutForms(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Fetch matches
	matches, err := server.Store().FetchMatches(event.Tag())
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/pdf")
	doc := pdf.New()
	renderMultipleScoutForms(doc, pdf.USLetterWidth, pdf.USLetterHeight, event, matches)
	return doc.Encode(w)
}
