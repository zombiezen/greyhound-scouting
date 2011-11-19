package main

import (
	"http"
	"os"
	"time"
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

func viewEvent(server *Server, w http.ResponseWriter, req *http.Request) os.Error {
	// TODO
	return nil
}
