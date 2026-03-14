package game

// State represents the current phase of a game.
type State int

const (
	GameStaging     State = 0 // Lobby
	RoundIntro      State = 1
	ShowQuestion    State = 2
	ShowAnswers     State = 3
	RevealTheTruth  State = 4
	ScoreBoard      State = 5
	ScoreBoardFinal State = 6
)

func (s State) String() string {
	switch s {
	case GameStaging:
		return "GameStaging"
	case RoundIntro:
		return "RoundIntro"
	case ShowQuestion:
		return "ShowQuestion"
	case ShowAnswers:
		return "ShowAnswers"
	case RevealTheTruth:
		return "RevealTheTruth"
	case ScoreBoard:
		return "ScoreBoard"
	case ScoreBoardFinal:
		return "ScoreBoardFinal"
	default:
		return "Unknown"
	}
}

// Duration returns the timer duration in milliseconds for timed states.
// Returns 0 for states with no timer.
func (s State) Duration() int {
	switch s {
	case RoundIntro:
		return 5000
	case ShowQuestion:
		return 25000
	case ShowAnswers:
		return 20000
	case ScoreBoard:
		return 5000
	default:
		return 0
	}
}
