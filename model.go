package main

import (
	"fmt"
	"launchpad.net/gobson/bson"
)

type Team struct {
	Number     int
	Name       string
	RookieYear int `bson:"rookie_year"`
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

func (t MatchType) Number() int {
	switch t {
	case Qualification:
		return 0
	case QuarterFinal:
		return 1
	case SemiFinal:
		return 2
	case Final:
		return 3
	}
	return -1
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

func (event *Event) Tag() string {
	return fmt.Sprintf("%s%04d", event.Location.Code, event.Date.Year)
}

type Alliance string

const (
	Red  Alliance = "red"
	Blue          = "blue"
)

func (alliance Alliance) String() string {
	return string(alliance)
}

type Match struct {
	Teams []TeamInfo
	Score map[Alliance]int `bson:",omitempty"`
	Type  MatchType
}

func MatchTag(event *Event, match *Match, matchNum int) string {
	return fmt.Sprintf("%s%d%03d", event.Tag(), match.Type.Number(), matchNum)
}

func MatchTeamTag(event *Event, match *Match, matchNum int, teamNum int) string {
	return MatchTag(event, match, matchNum) + fmt.Sprint(teamNum)
}

type TeamInfo struct {
	Team      int
	Alliance  Alliance
	Score     int
	ScoutName string `bson:"scout"`
	Failure   bool
	NoShow    bool
}
