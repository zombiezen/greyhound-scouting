package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

// A TagError is a tag parsing error.
type TagError struct {
	Tag     string
	BadPart string
	Err     error
}

func (e TagError) Error() string {
	if e.BadPart != "" {
		return fmt.Sprintf("Invalid tag %q: %v (at %q)", e.Tag, e.BadPart, e.Err)
	}
	return fmt.Sprintf("Invalid tag %q: %v", e.Tag, e.Err)
}

// wrapTagError returns a TagError containing e, or e if it is already a
// TagError.  The tag of the resulting error is always tag.  If e is nil, nil
// is returned.
func wrapTagError(e error, tag string) error {
	if e == nil {
		return e
	}
	if tagErr, ok := e.(TagError); ok {
		tagErr.Tag = tag
		return tagErr
	}
	return TagError{Tag: tag, Err: e}
}

type EventTag struct {
	LocationCode string
	Year         uint
}

func ParseEventTag(s string) (tag EventTag, err error) {
	defer func(tag string) {
		err = wrapTagError(err, tag)
	}(s)

	tag, s, err = parseEvent(s)
	if err == nil && s != "" {
		err = TagError{BadPart: s, Err: errors.New("Extra data at end of event tag")}
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
		err = TagError{Err: errors.New("Tag must begin with a location code")}
		return
	}
	if !isLowerString(tag.LocationCode) {
		err = TagError{BadPart: tag.LocationCode, Err: errors.New("Location code must be lowercase")}
		return
	}
	remaining = s[index:]

	// Parse 4-digit year
	if len(remaining) < yearWidth {
		err = TagError{BadPart: remaining, Err: fmt.Errorf("%d-digit year must follow location code", yearWidth)}
		return
	}
	year64, err := strconv.ParseUint(remaining[:yearWidth], 10, 0)
	if err != nil {
		err = TagError{BadPart: remaining[:yearWidth], Err: err}
		return
	}
	tag.Year = uint(year64)
	remaining = remaining[yearWidth:]

	return
}

// isLowerString returns true if and only if s contains only lowercase letters.
func isLowerString(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return !unicode.IsLower(r) }) == -1
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
	defer func(tag string) {
		err = wrapTagError(err, tag)
	}(s)

	tag.EventTag, s, err = parseEvent(s)
	if err != nil {
		return
	}
	tag.MatchType, tag.MatchNumber, s, err = parseMatch(s)
	if err == nil && s != "" {
		err = TagError{BadPart: s, Err: errors.New("Extra data at end of event tag")}
	}
	return
}

func parseMatch(s string) (matchType MatchType, matchNumber uint, remaining string, err error) {
	// Parse match type
	if len(s) == 0 {
		err = TagError{Err: errors.New("Missing one-digit match type")}
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
		err = TagError{BadPart: string(s[0]), Err: errors.New("Match type must be 0, 1, 2, or 3")}
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
	defer func(tag string) {
		err = wrapTagError(err, tag)
	}(s)

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
		err = TagError{BadPart: s, Err: errors.New("Extra data at end of match team tag")}
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
