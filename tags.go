package main

import (
	"fmt"
	"os"
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

func ParseEventTag(s string) (tag EventTag, err os.Error) {
	tag, s, err = parseEvent(s)
	if err == nil && s != "" {
		err = fmt.Errorf("Extra data at end of event tag: \"%s\"", s)
	}
	return
}

func parseEvent(s string) (tag EventTag, remaining string, err os.Error) {
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
		err = os.NewError("Tag must begin with a location code")
		return
	}
	remaining = s[index:]

	// Parse 4-digit year
	if len(remaining) < yearWidth {
		err = fmt.Errorf("%d-digit year must follow location code", yearWidth)
		return
	}
	if tag.Year, err = strconv.Atoui(remaining[:yearWidth]); err != nil {
		err = fmt.Errorf("%d-digit year must follow location code", yearWidth)
		return
	}
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

func ParseMatchTag(s string) (tag MatchTag, err os.Error) {
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

func parseMatch(s string) (matchType MatchType, matchNumber uint, remaining string, err os.Error) {
	// Parse match type
	if len(s) == 0 {
		err = os.NewError("Missing one-digit match type")
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
		err = os.NewError("Match type must be 0, 1, 2, or 3")
		return
	}

	// Parse match number
	if len(remaining) < matchNumberWidth {
		err = fmt.Errorf("Missing %d-digit match number", matchNumberWidth)
		return
	}
	matchNumberString, remaining := remaining[:matchNumberWidth], remaining[matchNumberWidth:]
	matchNumber, err2 := strconv.Atoui(matchNumberString)
	if err2 != nil {
		err = fmt.Errorf("Match number must be %d digits", matchNumberWidth)
		return
	}

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

func ParseMatchTeamTag(s string) (tag MatchTeamTag, err os.Error) {
	tag.EventTag, s, err = parseEvent(s)
	if err != nil {
		return
	}
	tag.MatchType, tag.MatchNumber, s, err = parseMatch(s)
	if err != nil {
		return
	}
	tag.TeamNumber, err = strconv.Atoui(s)
	if err != nil {
		err = fmt.Errorf("Extra data at end of match team tag: \"%s\"", s)
	}
	return
}

func (tag MatchTeamTag) String() string {
	return fmt.Sprintf("%s%0d", tag.MatchTag, tag.TeamNumber)
}

func (tag MatchTeamTag) GoString() string {
	return fmt.Sprintf("MatchTeamTag{MatchTag:%#v, TeamNumber:%d}", tag.MatchTag, tag.TeamNumber)
}
