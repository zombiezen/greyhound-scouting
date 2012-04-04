package main

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"code.google.com/p/gorilla/mux"
	"code.google.com/p/gorilla/schema"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"
)

func eventIndex(server *Server, w http.ResponseWriter, req *http.Request) error {
	// Query for events
	events := server.Store().Events(time.Now().Year())

	// Fetch events
	var eventList []Event
	if err := events.Limit(50).All(&eventList); err != nil {
		return err
	}

	// Render page
	return server.Templates().ExecuteTemplate(w, "event-index.html", map[string]interface{}{
		"Server":    server,
		"Request":   req,
		"EventList": eventList,
	})
}

func routeEventTag(vars map[string]string) EventTag {
	year64, _ := strconv.ParseUint(vars["year"], 10, 0)
	return EventTag{
		Year:         uint(year64),
		LocationCode: vars["location"],
	}
}

func viewEvent(server *Server, w http.ResponseWriter, req *http.Request) error {
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

	return server.Templates().ExecuteTemplate(w, "event.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Event":   event,
		"Matches": matches,
		"Teams":   teams,
	})
}

func routeMatchTag(vars map[string]string) MatchTag {
	num64, _ := strconv.ParseUint(vars["matchNumber"], 10, 0)
	return MatchTag{
		EventTag:    routeEventTag(vars),
		MatchType:   MatchType(vars["matchType"]),
		MatchNumber: uint(num64),
	}
}

func viewMatch(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Fetch match
	match, err := server.Store().FetchMatch(routeMatchTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	return server.Templates().ExecuteTemplate(w, "match.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Event":   event,
		"Match":   match,
	})
}

func editMatchTeam(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	var form struct {
		Autonomous   HoopCount
		Teleoperated HoopCount
		CoopBridge   Bridge
		TeamBridge1  Bridge
		TeamBridge2  Bridge
		ScoutName    string
	}

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Fetch match
	match, err := server.Store().FetchMatch(routeMatchTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Get team info
	teamNumber, _ := strconv.Atoi(vars["teamNumber"])
	var teamInfo *TeamInfo
	for i := range match.Teams {
		if match.Teams[i].Team == teamNumber {
			teamInfo = &match.Teams[i]
			break
		}
	}
	if teamInfo == nil {
		http.NotFound(w, req)
		return nil
	}

	// Parse forms
	if req.Method == "POST" {
		if err := req.ParseForm(); err != nil {
			// TODO: Bad request status code
			return err
		}

		d := schema.NewDecoder()
		if err := d.Decode(&form, req.Form); err != nil {
			return err
		}
		// TODO: Show errors in validation

		// Save
		teamInfo.Autonomous = form.Autonomous
		teamInfo.Teleoperated = form.Teleoperated
		teamInfo.CoopBridge = form.CoopBridge
		teamInfo.TeamBridge1 = form.TeamBridge1
		teamInfo.TeamBridge2 = form.TeamBridge2
		teamInfo.ScoutName = form.ScoutName
		if err := server.Store().UpdateMatchTeam(MatchTag{event.Tag(), match.Type, uint(match.Number)}, teamNumber, *teamInfo); err != nil {
			return err
		}

		// Redirect
		u, err := server.GetRoute("match.view").URL("year", strconv.Itoa(event.Date.Year), "location", event.Location.Code, "matchType", string(match.Type), "matchNumber", strconv.Itoa(match.Number))
		if err != nil {
			return err
		}
		http.Redirect(w, req, u.String(), http.StatusFound)
		return nil
	} else {
		form.Autonomous = teamInfo.Autonomous
		form.Teleoperated = teamInfo.Teleoperated
		form.CoopBridge = teamInfo.CoopBridge
		form.TeamBridge1 = teamInfo.TeamBridge1
		form.TeamBridge2 = teamInfo.TeamBridge2
		form.ScoutName = teamInfo.ScoutName
	}

	return server.Templates().ExecuteTemplate(w, "match-edit-team.html", map[string]interface{}{
		"Server":   server,
		"Request":  req,
		"Event":    event,
		"Match":    match,
		"TeamInfo": teamInfo,
		"Form":     form,
	})
}

func eventScoutForms(server *Server, w http.ResponseWriter, req *http.Request) error {
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

func matchSheet(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Fetch match
	match, err := server.Store().FetchMatch(routeMatchTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/pdf")
	doc := pdf.New()
	renderMatchSheet(doc, pdf.USLetterWidth, pdf.USLetterHeight, event, match)
	return doc.Encode(w)
}

func eventSpreadsheet(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Write header
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=teams.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{
		"Team #",
		"Matches Played",
		"No-Shows",
	})

	for _, teamNum := range event.Teams {
		// TODO: Get team stats
		cw.Write([]string{
			strconv.Itoa(teamNum),
			strconv.Itoa(0),
			strconv.Itoa(0),
		})
	}

	cw.Flush()
	return nil
}
