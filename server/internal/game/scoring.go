package game

// ScoreForCorrectAnswer returns points for selecting the real answer.
func ScoreForCorrectAnswer(roundIndex int) int {
	switch roundIndex {
	case 0:
		return 1000
	case 1:
		return 1500
	default:
		return 2000
	}
}

// ScoreForFoolingPlayer returns points earned per player fooled by your lie.
func ScoreForFoolingPlayer(roundIndex int) int {
	switch roundIndex {
	case 0:
		return 500
	case 1:
		return 750
	default:
		return 1000
	}
}

// ScoreForHouseLiePenalty returns the penalty for selecting a house lie (negative).
func ScoreForHouseLiePenalty(roundIndex int) int {
	switch roundIndex {
	case 0:
		return -500
	case 1:
		return -750
	default:
		return -1000
	}
}
