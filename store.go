package main

import (
	"os"
	"sort"

	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

var StoreNotFound = os.NewError("Not found in datastore")

// A Datastore can retrieve and store model objects.
type Datastore interface {
	Teams() Pager
	Events(year int) Pager

	FetchTeam(int) (*Team, os.Error)
	FetchEvent(EventTag) (*Event, os.Error)
	FetchMatches(EventTag) ([]*Match, os.Error)
	FetchMatch(MatchTag) (*Match, os.Error)
}

const (
	teamCollection  = "teams"
	eventCollection = "events"
)

// mongoDatastore persists model objects using MongoDB.
type mongoDatastore struct {
	mgo.Database
}

func (store mongoDatastore) Teams() Pager {
	return MongoPager{store.C(teamCollection).Find(nil).Sort(bson.D{{"_id", 1}})}
}

func (store mongoDatastore) Events(year int) Pager {
	return MongoPager{store.C(eventCollection).Find(bson.M{"date.year": year}).Sort(bson.D{{"date.month", 1}, {"date.day", 1}})}
}

func (store mongoDatastore) fetchOne(collection string, filter interface{}, ptr interface{}) os.Error {
	query := store.C(collection).Find(filter)
	err := query.One(ptr)
	if err == mgo.NotFound {
		err = StoreNotFound
	}
	return err
}

func (store mongoDatastore) FetchTeam(number int) (*Team, os.Error) {
	var team Team
	if err := store.fetchOne(teamCollection, bson.M{"_id": number}, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (store mongoDatastore) FetchEvent(tag EventTag) (*Event, os.Error) {
	var event Event
	if err := store.fetchOne(eventCollection, bson.M{"date.year": tag.Year, "location.code": tag.LocationCode}, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func matchCollection(tag EventTag) string {
	return "matches." + tag.String()
}

// Maximum number of matches to fetch per event.  Should be more than sufficient.
const matchLimit = 200

func (store mongoDatastore) FetchMatches(tag EventTag) ([]*Match, os.Error) {
	query := store.C(matchCollection(tag)).Find(nil).Limit(matchLimit)
	var matches []*Match
	if err := query.All(&matches); err != nil {
		return nil, err
	}
	sort.Sort(byMatchOrder(matches))
	return matches, nil
}

func (store mongoDatastore) FetchMatch(tag MatchTag) (*Match, os.Error) {
	var match Match
	if err := store.fetchOne(matchCollection(tag.EventTag), bson.M{"type": tag.MatchType, "number": tag.MatchNumber}, &match); err != nil {
		return nil, err
	}
	return &match, nil
}
