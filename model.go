package main

import (
	"sort"
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

func (t1 MatchType) Less(t2 MatchType) bool {
	return t1.key() < t2.key()
}

func (t MatchType) key() int {
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
	Score  map[string]int `bson:",omitempty"`
}

func (match *Match) AllianceInfo(alliance Alliance) AllianceInfo {
	// Get teams in alliance
	teams := make([]TeamInfo, 0, len(match.Teams)/2)
	for _, info := range match.Teams {
		if info.Alliance == alliance {
			teams = append(teams, info)
		}
	}
	sort.Sort(byTeamNumber(teams))

	// Get alliance score
	var score int
	if match.Score != nil {
		score = match.Score[string(alliance)]
	}

	// Create info struct
	return AllianceInfo{
		Alliance: alliance,
		Teams:    teams,
		Score:    score,
		Won:      match.Winner() == alliance,
	}
}

func (match *Match) Winner() Alliance {
	switch {
	case match.Score == nil:
		return ""
	case match.Score[string(Red)] > match.Score[string(Blue)]:
		return Red
	case match.Score[string(Red)] < match.Score[string(Blue)]:
		return Blue
	}
	return ""
}

type byMatchOrder []*Match

func (slice byMatchOrder) Len() int {
	return len(slice)
}

func (slice byMatchOrder) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice byMatchOrder) Less(i, j int) bool {
	if slice[i].Type.Less(slice[j].Type) {
		return true
	}
	return slice[i].Number < slice[j].Number
}

type TeamInfo struct {
	Team      int
	Alliance  Alliance
	Score     int
	ScoutName string `bson:"scout"`
	Failure   bool
	NoShow    bool
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

type AllianceInfo struct {
	Alliance Alliance
	Teams    []TeamInfo
	Score    int
	Won      bool
}
