package main

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	yearWidth        = 4
	matchNumberWidth = 3
)

const (
	qualificationDigit rune = '0' + iota
	quarterFinalDigit
	semiFinalDigit
	finalDigit
)

type EventTag struct {
	LocationCode string
	Year         uint
}

func ParseEventTag(s string) (tag EventTag, err error) {
	tag, s, err = parseEvent(s)
	if err == nil && s != "" {
		err = fmt.Errorf("Extra data at end of event tag: \"%s\"", s)
	}
	return
}

func parseEvent(s string) (tag EventTag, remaining string, err error) {
	// Find the first digit
	index := 0
	for ; index < len(s); index++ {
		if s[index] >= '0' && s[index] <= '9' {
			break
		}
	}

	// Ensure location is found
	tag.LocationCode = s[:index]
	if tag.LocationCode == "" {
		err = errors.New("Tag must begin with a location code")
		return
	}
	remaining = s[index:]

	// Parse 4-digit year
	if len(remaining) < yearWidth {
		err = fmt.Errorf("%d-digit year must follow location code", yearWidth)
		return
	}
	year64, err := strconv.ParseUint(remaining[:yearWidth], 10, 0)
	if err != nil {
		err = fmt.Errorf("%d-digit year must follow location code", yearWidth)
		return
	}
	tag.Year = uint(year64)
	remaining = remaining[yearWidth:]

	return
}

func (tag EventTag) String() string {
	return fmt.Sprintf("%s%0*d", tag.LocationCode, yearWidth, tag.Year)
}

func (tag EventTag) GoString() string {
	return fmt.Sprintf("EventTag{LocationCode:%q, Year:%d}", tag.LocationCode, tag.Year)
}

type MatchTag struct {
	EventTag
	MatchType   MatchType
	MatchNumber uint
}

func ParseMatchTag(s string) (tag MatchTag, err error) {
	tag.EventTag, s, err = parseEvent(s)
	if err != nil {
		return
	}
	tag.MatchType, tag.MatchNumber, s, err = parseMatch(s)
	if err == nil && s != "" {
		err = fmt.Errorf("Extra data at end of match tag: \"%s\"", s)
	}
	return
}

func parseMatch(s string) (matchType MatchType, matchNumber uint, remaining string, err error) {
	// Parse match type
	if len(s) == 0 {
		err = errors.New("Missing one-digit match type")
		return
	}
	digit, remaining := s[0], s[1:]
	switch rune(digit) {
	case qualificationDigit:
		matchType = Qualification
	case quarterFinalDigit:
		matchType = QuarterFinal
	case semiFinalDigit:
		matchType = SemiFinal
	case finalDigit:
		matchType = Final
	default:
		err = errors.New("Match type must be 0, 1, 2, or 3")
		return
	}

	// Parse match number
	if len(remaining) < matchNumberWidth {
		err = fmt.Errorf("Missing %d-digit match number", matchNumberWidth)
		return
	}
	matchNumberString, remaining := remaining[:matchNumberWidth], remaining[matchNumberWidth:]
	matchNumber64, err2 := strconv.ParseUint(matchNumberString, 10, 0)
	if err2 != nil {
		err = fmt.Errorf("Match number must be %d digits", matchNumberWidth)
		return
	}
	matchNumber = uint(matchNumber64)

	return
}

func (tag MatchTag) String() string {
	var typeDigit rune
	switch tag.MatchType {
	case Qualification:
		typeDigit = qualificationDigit
	case QuarterFinal:
		typeDigit = quarterFinalDigit
	case SemiFinal:
		typeDigit = semiFinalDigit
	case Final:
		typeDigit = finalDigit
	default:
		// This is an error.
		typeDigit = '!'
	}
	return fmt.Sprintf("%s%c%0*d", tag.EventTag.String(), typeDigit, matchNumberWidth, tag.MatchNumber)
}

func (tag MatchTag) GoString() string {
	return fmt.Sprintf("MatchTag{EventTag:%#v, MatchType:%q, MatchNumber:%d}", tag.EventTag, tag.MatchType, tag.MatchNumber)
}

type MatchTeamTag struct {
	MatchTag
	TeamNumber uint
}

func ParseMatchTeamTag(s string) (tag MatchTeamTag, err error) {
	tag.EventTag, s, err = parseEvent(s)
	if err != nil {
		return
	}
	tag.MatchType, tag.MatchNumber, s, err = parseMatch(s)
	if err != nil {
		return
	}
	teamNumber64, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		err = fmt.Errorf("Extra data at end of match team tag: \"%s\"", s)
	}
	tag.TeamNumber = uint(teamNumber64)
	return
}

func (tag MatchTeamTag) String() string {
	return fmt.Sprintf("%s%0d", tag.MatchTag, tag.TeamNumber)
}

func (tag MatchTeamTag) GoString() string {
	return fmt.Sprintf("MatchTeamTag{MatchTag:%#v, TeamNumber:%d}", tag.MatchTag, tag.TeamNumber)
}
