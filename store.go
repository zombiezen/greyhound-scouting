package main

import (
	"errors"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"sort"
)

var StoreNotFound = errors.New("Not found in datastore")

// A Datastore can retrieve and store model objects.
type Datastore interface {
	Teams() Pager
	Events(year int) Pager

	FetchTeam(int) (*Team, error)
	FetchTeams([]int) ([]*Team, error)
	FetchEvent(EventTag) (*Event, error)
	FetchMatches(EventTag) ([]*Match, error)
	FetchMatch(MatchTag) (*Match, error)

	EventsForTeam(year int, number int) ([]EventTag, error)

	TeamEventStats(EventTag, int) (TeamStats, error)

	UpsertTeam(Team) error
	UpdateMatchTeam(MatchTag, int, TeamInfo) error
}

const (
	teamCollection  = "teams"
	eventCollection = "events"
)

// mongoDatastore persists model objects using MongoDB.
type mongoDatastore struct {
	*mgo.Database
}

func (store mongoDatastore) Teams() Pager {
	return MongoPager{store.C(teamCollection).Find(nil).Sort(bson.D{{"_id", 1}})}
}

func (store mongoDatastore) Events(year int) Pager {
	return MongoPager{store.C(eventCollection).Find(bson.M{"date.year": year}).Sort(bson.D{{"date.month", 1}, {"date.day", 1}})}
}

func (store mongoDatastore) fetchOne(collection string, filter interface{}, ptr interface{}) error {
	query := store.C(collection).Find(filter)
	err := query.One(ptr)
	if err == mgo.NotFound {
		err = StoreNotFound
	}
	return err
}

func (store mongoDatastore) FetchTeams(numbers []int) ([]*Team, error) {
	query := store.C(teamCollection).Find(bson.M{"_id": bson.M{"$in": numbers}}).Sort(bson.D{{"_id", 1}})
	var teams []*Team
	if err := query.All(&teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func (store mongoDatastore) FetchTeam(number int) (*Team, error) {
	var team Team
	if err := store.fetchOne(teamCollection, bson.M{"_id": number}, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (store mongoDatastore) FetchEvent(tag EventTag) (*Event, error) {
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

func (store mongoDatastore) FetchMatches(tag EventTag) ([]*Match, error) {
	query := store.C(matchCollection(tag)).Find(nil).Limit(matchLimit)
	var matches []*Match
	if err := query.All(&matches); err != nil {
		return nil, err
	}
	sort.Sort(byMatchOrder(matches))
	return matches, nil
}

func (store mongoDatastore) FetchMatch(tag MatchTag) (*Match, error) {
	var match Match
	if err := store.fetchOne(matchCollection(tag.EventTag), bson.M{"type": tag.MatchType, "number": tag.MatchNumber}, &match); err != nil {
		return nil, err
	}
	return &match, nil
}

func (store mongoDatastore) EventsForTeam(year int, number int) ([]EventTag, error) {
	query := store.C(eventCollection).Find(bson.M{"date.year": year, "teams": number}).Sort(bson.D{{"date.month", 1}, {"date.day", 1}})
	var events []Event
	if err := query.All(&events); err != nil {
		return nil, err
	}
	tags := make([]EventTag, len(events))
	for i := range events {
		tags[i] = events[i].Tag()
	}
	return tags, nil
}

// TeamEventStats returns team statistics for a single event.
func (store mongoDatastore) TeamEventStats(tag EventTag, number int) (TeamStats, error) {
	iter := store.C(matchCollection(tag)).Find(bson.M{"teams.team": number}).Limit(matchLimit).Iter()

	var stats TeamStats
	var match Match
	stats.EventTag = tag
	for iter.Next(&match) {
		var i int
		for i = 0; i < len(match.Teams); i++ {
			if match.Teams[i].Team == number {
				break
			}
		}
		if i >= len(match.Teams) {
			// Team not found in match.  This shouldn't be hit.
			// TODO: Log problem
			continue
		}

		if match.Teams[i].NoShow {
			stats.NoShowCount++
			continue
		}

		if match.Score == nil {
			continue
		}

		stats.MatchCount++
		stats.TotalPoints += match.Teams[i].Score
		if match.Teams[i].Failure {
			stats.Failures++
		}
		stats.AutonomousHoops.Add(match.Teams[i].Autonomous)
		stats.TeleoperatedHoops.Add(match.Teams[i].Teleoperated)
	}
	return stats, iter.Err()
}

func (store mongoDatastore) UpsertTeam(team Team) error {
	_, err := store.C(teamCollection).Upsert(bson.M{"_id": team.Number}, team)
	return err
}

func (store mongoDatastore) UpdateMatchTeam(tag MatchTag, teamNumber int, info TeamInfo) error {
	return store.C(matchCollection(tag.EventTag)).Update(
		bson.M{"type": tag.MatchType, "number": tag.MatchNumber, "teams.team": teamNumber},
		bson.M{"$set": bson.M{"teams.$": info}},
	)
}
