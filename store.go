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

	TeamEventMatches(EventTag, int) ([]*Match, error)
	TeamEventStats(EventTag, int) (TeamStats, error)

	UpdateMatchScore(MatchTag, int, int) error
	UpdateMatchTeam(MatchTag, int, TeamInfo) error

	UpsertTeam(*Team) error
	UpsertEvent(*Event) error
	UpsertMatch(EventTag, *Match) error
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

func (store mongoDatastore) TeamEventMatches(tag EventTag, number int) ([]*Match, error) {
	query := store.C(matchCollection(tag)).Find(bson.M{"teams.team": number}).Limit(matchLimit)
	var matches []*Match
	if err := query.All(&matches); err != nil {
		return nil, err
	}
	sort.Sort(byMatchOrder(matches))
	return matches, nil
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
			stats.FailureCount++
		}
		stats.AutonomousHoops.Add(match.Teams[i].Autonomous)
		stats.TeleoperatedHoops.Add(match.Teams[i].Teleoperated)
		stats.CoopBridge.add(match.Teams[i].CoopBridge)
		stats.TeamBridge1.add(match.Teams[i].TeamBridge1)
		stats.TeamBridge2.add(match.Teams[i].TeamBridge2)
	}
	return stats, iter.Err()
}

func (store mongoDatastore) UpsertTeam(team *Team) error {
	_, err := store.C(teamCollection).Upsert(bson.M{"_id": team.Number}, team)
	return err
}

func (store mongoDatastore) UpsertEvent(event *Event) error {
	_, err := store.C(eventCollection).Upsert(bson.M{"location.code": event.Location.Code, "date.year": event.Date.Year}, event)
	return err
}

func (store mongoDatastore) UpsertMatch(etag EventTag, match *Match) error {
	_, err := store.C(matchCollection(etag)).Upsert(bson.M{"type": match.Type, "number": match.Number}, match)
	return err
}

func (store mongoDatastore) UpdateMatchScore(tag MatchTag, red int, blue int) error {
	return store.C(matchCollection(tag.EventTag)).Update(
		bson.M{"type": tag.MatchType, "number": tag.MatchNumber},
		bson.M{"$set": bson.M{"score.red": red, "score.blue": blue}},
	)
}

func (store mongoDatastore) UpdateMatchTeam(tag MatchTag, teamNumber int, info TeamInfo) error {
	return store.C(matchCollection(tag.EventTag)).Update(
		bson.M{"type": tag.MatchType, "number": tag.MatchNumber, "teams.team": teamNumber},
		bson.M{"$set": bson.M{"teams.$": info}},
	)
}
