package main

import (
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

type Match struct {
	Teams []TeamInfo
	Score map[Alliance]int `bson:",omitempty"`
	Type  MatchType
}

type TeamInfo struct {
	Team      int
	Alliance  Alliance
	Score     int
	ScoutName string `bson:"scout"`
	Failure   bool
	NoShow    bool
}
