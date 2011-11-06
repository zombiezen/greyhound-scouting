package main

import (
	"http"
	"os"
	"strconv"

	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"

	"gorilla.googlecode.com/hg/gorilla/mux"
)

func viewTeam(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	vars := mux.Vars(req)
	number, _ := strconv.Atoi(vars["number"])

	// Fetch team
	var team Team
	err := server.DB().C("teams").Find(bson.M{"number": number}).One(&team)
	if err == mgo.NotFound {
		http.NotFound(w, req)
		return nil
	} else if err != nil {
		return err
	}

	// TODO: stats
	// TODO: image

	server.TemplateSet().Execute(w, "team.html", map[string]interface{}{
		"Server":  server,
		"Request": req,
		"Team":    team,
	})
	return nil
}
