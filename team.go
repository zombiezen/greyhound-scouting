package main

import (
	"http"
	"os"
	"strconv"

	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

func teamIndex(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	// Determine page number
	pageNumber, err := strconv.Atoi(req.FormValue("page"))
	if err != nil {
		pageNumber = 1
	}

	// Query for teams
	teams := server.DB().C("teams").Find(nil).Sort(bson.D{{"_id", 1}})

	// Paginate teams
	p, err := NewPaginator(MongoPager{teams}, 50)
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
	return server.TemplateSet().Execute(w, "team-index.html", map[string]interface{}{
		"Server":   server,
		"Request":  req,
		"TeamList": teamList,
		"Page":     page,
	})
}

func viewTeam(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)
	number, _ := strconv.Atoi(vars["number"])

	// Fetch team
	var team Team
	err := server.DB().C("teams").Find(bson.M{"_id": number}).One(&team)
	if err == mgo.NotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// TODO: stats
	// TODO: image

	return server.TemplateSet().Execute(w, "team.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Team":    team,
	})
}
