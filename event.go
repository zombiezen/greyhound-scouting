package main

import (
	"http"
	"os"
	"time"

	"launchpad.net/gobson/bson"
)

func eventIndex(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	// Query for events
	currentSeason := time.LocalTime().Year
	events := server.DB().C("events").Find(bson.M{"date.year": currentSeason}).Sort(bson.D{{"date.month", 1}, {"date.day", 1}})

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

func viewEvent(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	// TODO
	return nil
}
