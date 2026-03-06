package game

import (
	"testing"
)

func TestGameStateDuration(t *testing.T) {
	tests := []struct {
		state    GameState
		expected int
	}{
		{GameStaging, 0},
		{RoundIntro, 5000},
		{ShowQuestion, 25000},
		{ShowAnswers, 20000},
		{RevealTheTruth, 0},
		{ScoreBoard, 5000},
		{ScoreBoardFinal, 0},
	}
	for _, tt := range tests {
		got := tt.state.Duration()
		if got != tt.expected {
			t.Errorf("%s.Duration() = %d, want %d", tt.state, got, tt.expected)
		}
	}
}
