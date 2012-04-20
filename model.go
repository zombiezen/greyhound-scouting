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
	Blue Alliance = "blue"
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

// AlliancePairs returns pairs of team infos.
func (match *Match) AlliancePairs() []struct{ Red, Blue *TeamInfo } {
	red := match.AllianceInfo(Red)
	blue := match.AllianceInfo(Blue)
	maxLen := len(red.Teams)
	if len(blue.Teams) > maxLen {
		maxLen = len(blue.Teams)
	}
	pairs := make([]struct{ Red, Blue *TeamInfo }, maxLen)
	for i := range pairs {
		if i < len(red.Teams) {
			pairs[i].Red = &red.Teams[i]
		}
		if i < len(blue.Teams) {
			pairs[i].Blue = &blue.Teams[i]
		}
	}
	return pairs
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

// TeamInfo returns the team info for a particular team.
func (match *Match) TeamInfo(teamNum int) *TeamInfo {
	for i := range match.Teams {
		if match.Teams[i].Team == teamNum {
			return &match.Teams[i]
		}
	}
	return nil
}

type byMatchOrder []*Match

func (slice byMatchOrder) Len() int {
	return len(slice)
}

func (slice byMatchOrder) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice byMatchOrder) Less(i, j int) bool {
	if slice[i].Type != slice[j].Type {
		return slice[i].Type.Less(slice[j].Type)
	}
	return slice[i].Number < slice[j].Number
}

type TeamInfo struct {
	Team     int
	Alliance Alliance
	Score    int

	ScoutName    string `bson:"scout"`
	Autonomous   HoopCount
	Teleoperated HoopCount
	CoopBridge   Bridge
	TeamBridge1  Bridge
	TeamBridge2  Bridge

	// These currently won't be used.
	Failure bool
	NoShow  bool
}

type Hoop int

const (
	NoHoop Hoop = iota
	HighHoop
	MidHoop
	LowHoop
)

// HoopCount stores how many hoops were scored per robot per phase.
type HoopCount struct {
	High int
	Mid  int
	Low  int
}

// score returns the score for a hoop count for the high, mid, and low score multipliers.
func (h HoopCount) score(high, mid, low int) int {
	return h.High*high + h.Mid*mid + h.Low*low
}

// Total returns the total number of hoops scored.
func (h HoopCount) Total() int {
	return h.High + h.Mid + h.Low
}

// Max returns the hoop with the maximum count.  If all hoops are zero, then NoHoop is returned.
func (h HoopCount) Max() Hoop {
	switch {
	case h.High == 0 && h.Mid == 0 && h.Low == 0:
		return NoHoop
	case h.Low > h.Mid && h.Low > h.High:
		return LowHoop
	case h.Mid > h.Low && h.Mid > h.High:
		return MidHoop
	}
	return HighHoop
}

// Add increments h1 by h2.
func (h1 *HoopCount) Add(h2 HoopCount) {
	h1.High += h2.High
	h1.Mid += h2.Mid
	h1.Low += h2.Low
}

// Bridge stores a match bridge attempt.
type Bridge struct {
	Attempted bool
	Success   bool
}

// CalculateScore computes a team's score.
func CalculateScore(auto, teleop HoopCount, coop, bridge1, bridge2 Bridge) int {
	const (
		teleopHighPoints = 3
		teleopMidPoints  = 2
		teleopLowPoints  = 1

		autoHighPoints = teleopHighPoints + 3
		autoMidPoints  = teleopMidPoints + 3
		autoLowPoints  = teleopLowPoints + 3

		bridgePoints = 10
	)

	autoScore := auto.score(autoHighPoints, autoMidPoints, autoLowPoints)
	teleopScore := teleop.score(teleopHighPoints, teleopMidPoints, teleopLowPoints)
	bridgeScore := 0
	if bridge1.Success {
		bridgeScore += bridgePoints
	}

	return autoScore + teleopScore + bridgeScore
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

// TeamStats holds team statistics.
type TeamStats struct {
	EventTag    EventTag
	MatchCount  int
	TotalPoints int

	NoShowCount  int
	FailureCount int

	CoopBridge  BridgeStats
	TeamBridge1 BridgeStats
	TeamBridge2 BridgeStats

	AutonomousHoops   HoopCount
	TeleoperatedHoops HoopCount
}

// AverageScore returns the average score.  Returns 0.0 if match count is zero.
func (stats TeamStats) AverageScore() float64 {
	if stats.MatchCount == 0 {
		return 0.0
	}
	return float64(stats.TotalPoints) / float64(stats.MatchCount)
}

// FailureRate returns the number of failures divided by the number of matches played.  Returns 0.0 if match count is zero.
func (stats TeamStats) FailureRate() float64 {
	if stats.MatchCount == 0 {
		return 0.0
	}
	return float64(stats.FailureCount) / float64(stats.MatchCount)
}

// AverageTeleoperatedHoops returns the average number of hoops scored per match.  Returns 0.0 if match count is zero.
func (stats TeamStats) AverageTeleoperatedHoops() float64 {
	if stats.MatchCount == 0 {
		return 0.0
	}
	return float64(stats.TeleoperatedHoops.Total()) / float64(stats.MatchCount)
}

// AverageAutonomousHoops returns the average number of hoops scored per match.  Returns 0.0 if match count is zero.
func (stats TeamStats) AverageAutonomousHoops() float64 {
	if stats.MatchCount == 0 {
		return 0.0
	}
	return float64(stats.AutonomousHoops.Total()) / float64(stats.MatchCount)
}

// BridgeStats holds team statistics for a particular bridge.
type BridgeStats struct {
	AttemptCount int
	SuccessCount int
}

// AttemptRate returns the number of attempts divided by matchCount.  Returns 0.0 if matchCount is zero.
func (stats BridgeStats) AttemptRate(matchCount int) float64 {
	if matchCount == 0 {
		return 0.0
	}
	return float64(stats.AttemptCount) / float64(matchCount)
}

// SuccessRate returns the number of successes divided by the number of attempts.  Returns 0.0 if number of attempts is zero.
func (stats BridgeStats) SuccessRate() float64 {
	if stats.AttemptCount == 0 {
		return 0.0
	}
	return float64(stats.SuccessCount) / float64(stats.AttemptCount)
}

// addBridge adds a single bridge to stats.
func (stats *BridgeStats) add(b Bridge) {
	if b.Attempted {
		stats.AttemptCount++
	}
	if b.Success {
		stats.SuccessCount++
	}
}
