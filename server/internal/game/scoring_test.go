package game

import (
	"testing"
)

func TestScoreForCorrectAnswer(t *testing.T) {
	tests := []struct {
		round    int
		expected int
	}{
		{0, 1000},
		{1, 1500},
		{2, 2000},
	}
	for _, tt := range tests {
		got := ScoreForCorrectAnswer(tt.round)
		if got != tt.expected {
			t.Errorf("ScoreForCorrectAnswer(%d) = %d, want %d", tt.round, got, tt.expected)
		}
	}
}

func TestScoreForFoolingPlayer(t *testing.T) {
	tests := []struct {
		round    int
		expected int
	}{
		{0, 500},
		{1, 750},
		{2, 1000},
	}
	for _, tt := range tests {
		got := ScoreForFoolingPlayer(tt.round)
		if got != tt.expected {
			t.Errorf("ScoreForFoolingPlayer(%d) = %d, want %d", tt.round, got, tt.expected)
		}
	}
}

func TestScoreForHouseLiePenalty(t *testing.T) {
	tests := []struct {
		round    int
		expected int
	}{
		{0, -500},
		{1, -750},
		{2, -1000},
	}
	for _, tt := range tests {
		got := ScoreForHouseLiePenalty(tt.round)
		if got != tt.expected {
			t.Errorf("ScoreForHouseLiePenalty(%d) = %d, want %d", tt.round, got, tt.expected)
		}
	}
}
