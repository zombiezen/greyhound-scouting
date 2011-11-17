package main

import (
	"launchpad.net/gobson/bson"
	"sort"
)

type Team struct {
	Number     int
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
	QuarterFinal            = "quarter"
	SemiFinal               = "semifinal"
	Final                   = "final"
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
	MatchIDs map[MatchType][]bson.ObjectId `bson:"matches"`
	Teams    []int
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
	Teams []TeamInfo
	Score map[Alliance]int `bson:",omitempty"`
	Type  MatchType
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
	teams := make([]TeamInfo, 0, len(match.Teams) / 2)
	for _, info := range match.Teams {
		if info.Alliance == alliance {
			teams = append(teams, info)
		}
	}
	sort.Sort(byTeamNumber(teams))
	return teams
}

type NumberedMatch struct {
	Number int
	Match
}

type TeamInfo struct {
	Team      int
	Alliance  Alliance
	Score     int
	ScoutName string `bson:"scout"`
	Failure   bool
	NoShow    bool
}
