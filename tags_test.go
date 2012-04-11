package main

import (
	"fmt"
	"reflect"
	"testing"
)

type tagTest struct {
	String string
	Tag    interface{}
}

func checkTagParseResult(t *testing.T, name string, tt tagTest, result interface{}, err error) {
	switch {
	case err != nil && tt.Tag != nil:
		t.Errorf("%s(%q) error: %v", name, tt.String, err)
	case err == nil && tt.Tag == nil:
		t.Errorf("%s(%q) did not produce an error", name, tt.String)
	case err == nil && !reflect.DeepEqual(result, tt.Tag):
		t.Errorf("%s(%q) != %#v (got %#v)", name, tt.String, tt.Tag, result)
	}
}

func checkTagString(t *testing.T, tt tagTest) {
	result := tt.Tag.(fmt.Stringer).String()
	if result != tt.String {
		t.Errorf("%#v.String() != %q (got %q)", tt.Tag, tt.String, result)
	}
}

var eventTagTests = []tagTest{
	{"sdc2011", EventTag{"sdc", 2011}},
	{"sdc0008", EventTag{"sdc", 8}},
	{"2011", nil},
	{"sdc201", nil},
	{"sdc2a11", nil},
	{"sdc2011a", nil},
	{"SDC2011", nil},
}

func TestParseEventTag(t *testing.T) {
	for _, tt := range eventTagTests {
		result, err := ParseEventTag(tt.String)
		checkTagParseResult(t, "ParseEventTag", tt, result, err)
	}
}

func TestEventTagString(t *testing.T) {
	for _, tt := range eventTagTests {
		if tt.Tag != nil {
			checkTagString(t, tt)
		}
	}
}

var matchTagTests = []tagTest{
	{"sdc20110042", MatchTag{EventTag{"sdc", 2011}, Qualification, 42}},
	{"sdc20111042", MatchTag{EventTag{"sdc", 2011}, QuarterFinal, 42}},
	{"sdc20112042", MatchTag{EventTag{"sdc", 2011}, SemiFinal, 42}},
	{"sdc20113042", MatchTag{EventTag{"sdc", 2011}, Final, 42}},
	{"sdc20114042", nil},
	{"20110042", nil},
	{"sdc201100421", nil},
	{"sdc20110042a", nil},
	{"sdc201100a2", nil},
	{"SDC20113042", nil},
}

func TestParseMatchTag(t *testing.T) {
	for _, tt := range matchTagTests {
		result, err := ParseMatchTag(tt.String)
		checkTagParseResult(t, "ParseMatchTag", tt, result, err)
	}
}

func TestMatchTagString(t *testing.T) {
	for _, tt := range matchTagTests {
		if tt.Tag != nil {
			checkTagString(t, tt)
		}
	}
}

var matchTeamTagTests = []tagTest{
	{"sdc201100421", MatchTeamTag{MatchTag{EventTag{"sdc", 2011}, Qualification, 42}, 1}},
	{"sdc20110042973", MatchTeamTag{MatchTag{EventTag{"sdc", 2011}, Qualification, 42}, 973}},
	{"sdc201100421a", nil},
	{"SDC201100421", nil},
}

func TestParseMatchTeamTag(t *testing.T) {
	for _, tt := range matchTeamTagTests {
		result, err := ParseMatchTeamTag(tt.String)
		checkTagParseResult(t, "ParseMatchTeamTag", tt, result, err)
	}
}

func TestMatchTeamTagString(t *testing.T) {
	for _, tt := range matchTeamTagTests {
		if tt.Tag != nil {
			checkTagString(t, tt)
		}
	}
}
