package main

import (
	"code.google.com/p/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func teamIndex(server *Server, w http.ResponseWriter, req *http.Request) error {
	// Determine page number
	pageNumber, err := strconv.Atoi(req.FormValue("page"))
	if err != nil {
		pageNumber = 1
	}

	// Paginate teams
	p, err := NewPaginator(server.Store().Teams(), 50)
	if err != nil {
		return err
	}
	page := p.Page(pageNumber)
	if page == nil {
		http.NotFound(w, req)
		return nil
	}

	// Get team list
	var teamList []Team
	if err := page.Get(&teamList); err != nil {
		return err
	}

	// Render page
	return server.Templates().ExecuteTemplate(w, "team-index.html", map[string]interface{}{
		"Server":   server,
		"Request":  req,
		"TeamList": teamList,
		"Page":     page,
	})
}

func viewTeam(server *Server, w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	number, _ := strconv.Atoi(vars["number"])

	// Fetch team
	team, err := server.Store().FetchTeam(number)
	if err == StoreNotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// Stats
	eventTags, err := server.Store().EventsForTeam(time.Now().Year(), number)
	if err != nil {
		return err
	}
	stats := make([]TeamStats, len(eventTags))
	for i := range eventTags {
		stats[i], err = server.Store().TeamEventStats(eventTags[i], number)
		if err != nil {
			return err
		}
	}
	// TODO: image

	return server.Templates().ExecuteTemplate(w, "team.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Team":    team,
		"Stats":   stats,
	})
}
