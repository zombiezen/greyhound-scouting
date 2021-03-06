package main

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"code.google.com/p/gorilla/mux"
	"code.google.com/p/gorilla/schema"
	"encoding/csv"
	"log"
	"net/http"
	"reflect"
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

func teamMatches(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	// Fetch event
	event, err := server.Store().FetchEvent(routeEventTag(vars))
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Ensure team is at event
	teamNumber, _ := strconv.Atoi(vars["teamNumber"])
	foundTeam := false
	for _, t := range event.Teams {
		if teamNumber == t {
			foundTeam = true
			break
		}
	}
	if !foundTeam {
		http.NotFound(w, req)
		return nil
	}

	// Fetch matches
	matches, err := server.Store().TeamEventMatches(event.Tag(), teamNumber)
	if err != nil {
		return err
	}

	return server.Templates().ExecuteTemplate(w, "team-matches.html", map[string]interface{}{
		"Server":     server,
		"Request":    req,
		"Event":      event,
		"TeamNumber": teamNumber,
		"Matches":    matches,
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

func scoreMatch(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	var form struct {
		RedScore  int
		BlueScore int
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
		if err := server.Store().UpdateMatchScore(MatchTag{event.Tag(), match.Type, uint(match.Number)}, form.RedScore, form.BlueScore); err != nil {
			return err
		}
	}

	// Redirect
	u, err := server.GetRoute("match.view").URL("year", strconv.Itoa(event.Date.Year), "location", event.Location.Code, "matchType", string(match.Type), "matchNumber", strconv.Itoa(match.Number))
	if err != nil {
		return err
	}
	http.Redirect(w, req, u.String(), http.StatusFound)
	return nil
}

func convertBridgeString(s string) reflect.Value {
	switch s {
	case "na":
		return reflect.ValueOf(Bridge{false, false})
	case "fail":
		return reflect.ValueOf(Bridge{true, false})
	case "success":
		return reflect.ValueOf(Bridge{true, true})
	}
	return reflect.Value{}
}

func editMatchTeam(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	var form struct {
		Autonomous   BallCount
		Teleoperated BallCount
		CoopBridge   Bridge
		TeamBridge1  Bridge
		TeamBridge2  Bridge
		ScoutName    string
		Failure      bool
		NoShow       bool
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
		d.RegisterConverter(Bridge{}, convertBridgeString)
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
		teamInfo.Failure = form.Failure
		teamInfo.NoShow = form.NoShow
		teamInfo.Score = CalculateScore(teamInfo.Autonomous, teamInfo.Teleoperated, teamInfo.CoopBridge, teamInfo.TeamBridge1, teamInfo.TeamBridge2)
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
		form.Failure = teamInfo.Failure
		form.NoShow = teamInfo.NoShow
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
	renderMatchSheet(doc, pdf.USLetterHeight, pdf.USLetterWidth, event, match, server.Store(), server.imagestore)
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
		"Failures",
		"Average Score",
		"Average Teleop Scored",
		"Average Teleop Shot",
		"Average Auto Scored",
		"Average Auto Shot",
		"Max Teleop Scored",
		"Max Teleop Shot",
		"Coop Bridge Attempts",
		"Coop Bridge Successes",
		"Bridge 1 Attempts",
		"Bridge 1 Successes",
		"Bridge 2 Attempts",
		"Bridge 2 Successes",
		"Auto High",
		"Auto Mid",
		"Auto Low",
		"Auto Missed",
		"Teleop High",
		"Teleop Mid",
		"Teleop Low",
		"Teleop Missed",
	})

	for _, teamNum := range event.Teams {
		stats, err := server.Store().TeamEventStats(event.Tag(), teamNum)
		if err != nil {
			log.Printf("Stats failed for team %d: %v", teamNum, err)
		}
		cw.Write([]string{
			strconv.Itoa(teamNum),
			strconv.Itoa(stats.MatchCount),
			strconv.Itoa(stats.NoShowCount),
			strconv.Itoa(stats.FailureCount),
			strconv.FormatFloat(stats.AverageScore(), 'f', -1, 64),
			strconv.FormatFloat(stats.AverageTeleoperatedScored(), 'f', -1, 64),
			strconv.FormatFloat(stats.AverageTeleoperatedShot(), 'f', -1, 64),
			strconv.FormatFloat(stats.AverageAutonomousScored(), 'f', -1, 64),
			strconv.FormatFloat(stats.AverageAutonomousShot(), 'f', -1, 64),
			strconv.Itoa(stats.MaxTeleoperatedScored),
			strconv.Itoa(stats.MaxTeleoperatedShot),

			strconv.Itoa(stats.CoopBridge.AttemptCount),
			strconv.Itoa(stats.CoopBridge.SuccessCount),
			strconv.Itoa(stats.TeamBridge1.AttemptCount),
			strconv.Itoa(stats.TeamBridge1.SuccessCount),
			strconv.Itoa(stats.TeamBridge2.AttemptCount),
			strconv.Itoa(stats.TeamBridge2.SuccessCount),

			strconv.Itoa(stats.AutonomousBalls.High),
			strconv.Itoa(stats.AutonomousBalls.Mid),
			strconv.Itoa(stats.AutonomousBalls.Low),
			strconv.Itoa(stats.AutonomousBalls.Missed),
			strconv.Itoa(stats.TeleoperatedBalls.High),
			strconv.Itoa(stats.TeleoperatedBalls.Mid),
			strconv.Itoa(stats.TeleoperatedBalls.Low),
			strconv.Itoa(stats.TeleoperatedBalls.Missed),
		})
	}

	cw.Flush()
	return nil
}
