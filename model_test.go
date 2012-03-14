package main

import (
	"testing"
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		Autonomous   HoopCount
		Teleoperated HoopCount
		Coop         Bridge
		Bridge1      Bridge
		Bridge2      Bridge
		Expected     int
	}{
		{
			Expected: 0,
		},
		{
			Expected:     1,
			Autonomous:   HoopCount{0, 0, 0},
			Teleoperated: HoopCount{0, 0, 1},
			Coop:         Bridge{false, false},
			Bridge1:      Bridge{false, false},
			Bridge2:      Bridge{false, false},
		},
		{
			Expected:     4,
			Autonomous:   HoopCount{0, 0, 1},
			Teleoperated: HoopCount{0, 0, 0},
			Coop:         Bridge{false, false},
			Bridge1:      Bridge{false, false},
			Bridge2:      Bridge{false, false},
		},
		// TODO: More hoop tests
		{
			Expected:     0,
			Autonomous:   HoopCount{0, 0, 0},
			Teleoperated: HoopCount{0, 0, 0},
			Coop:         Bridge{false, false},
			Bridge1:      Bridge{true, false},
			Bridge2:      Bridge{false, false},
		},
		{
			Expected:     10,
			Autonomous:   HoopCount{0, 0, 0},
			Teleoperated: HoopCount{0, 0, 0},
			Coop:         Bridge{false, false},
			Bridge1:      Bridge{true, true},
			Bridge2:      Bridge{false, false},
		},
		{
			Expected:     10,
			Autonomous:   HoopCount{0, 0, 0},
			Teleoperated: HoopCount{0, 0, 0},
			Coop:         Bridge{false, false},
			Bridge1:      Bridge{true, true},
			Bridge2:      Bridge{true, true},
		},
	}
	for _, tt := range tests {
		result := CalculateScore(tt.Autonomous, tt.Teleoperated, tt.Coop, tt.Bridge1, tt.Bridge2)
		if result != tt.Expected {
			t.Errorf("CalculateScore(%v, %v, %v, %v, %v) != %d (got %d)",
				tt.Autonomous, tt.Teleoperated, tt.Coop, tt.Bridge1, tt.Bridge2,
				tt.Expected, result)
		}
	}
}
