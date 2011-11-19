package main

import (
	"os"
	"sort"

	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

type Team struct {
	Number     int `bson:"_id"`
	Name       string
	RookieYear int `bson:"rookie_year"`
	Robot      *Robot
}

type Robot struct {
	Name  string
	Image string
	Notes string `bson:",omitempty"`
}

type MatchType string

const (
	Qualification MatchType = "qualification"
	QuarterFinal  MatchType = "quarter"
	SemiFinal     MatchType = "semifinal"
	Final         MatchType = "final"
)

func (t MatchType) String() string {
	return string(t)
}

func (t MatchType) DisplayName() string {
	switch t {
	case Qualification:
		return "Qualification"
	case QuarterFinal:
		return "Quarter-Final"
	case SemiFinal:
		return "Semi-Final"
	case Final:
		return "Final"
	}
	return string(t)
}

type Event struct {
	Location struct {
		Name string
		Code string
	}
	Date struct {
		Year  int
		Month int
		Day   int
	}
	Teams []int
}

func (event *Event) Tag() EventTag {
	return EventTag{
		LocationCode: event.Location.Code,
		Year:         uint(event.Date.Year),
	}
}

type Alliance string

const (
	Red  Alliance = "red"
	Blue          = "blue"
)

func (alliance Alliance) String() string {
	return string(alliance)
}

func (alliance Alliance) DisplayName() string {
	switch alliance {
	case Red:
		return "Red"
	case Blue:
		return "Blue"
	}
	return string(alliance)
}

type Match struct {
	Type   MatchType
	Number int
	Teams  []TeamInfo
	Score  map[Alliance]int `bson:",omitempty"`
}

type byTeamNumber []TeamInfo

func (slice byTeamNumber) Len() int {
	return len(slice)
}

func (slice byTeamNumber) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice byTeamNumber) Less(i, j int) bool {
	return slice[i].Team < slice[j].Team
}

func (match *Match) Alliance(alliance Alliance) []TeamInfo {
	teams := make([]TeamInfo, 0, len(match.Teams)/2)
	for _, info := range match.Teams {
		if info.Alliance == alliance {
			teams = append(teams, info)
		}
	}
	sort.Sort(byTeamNumber(teams))
	return teams
}

type TeamInfo struct {
	Team      int
	Alliance  Alliance
	Score     int
	ScoutName string `bson:"scout"`
	Failure   bool
	NoShow    bool
}

func FetchEvent(database mgo.Database, tag EventTag) (*Event, os.Error) {
	query := database.C("events").Find(bson.M{"date.year": tag.Year, "location.code": tag.LocationCode})
	var event Event
	if err := query.One(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func matchCollection(tag EventTag) string {
	return "matches." + tag.String()
}

func FetchMatches(database mgo.Database, eventTag EventTag) ([]*Match, os.Error) {
	// TODO
	return nil, nil
}
