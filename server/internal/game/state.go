package game

// GameState represents the current phase of a game.
type GameState int

const (
	GameStaging     GameState = 0 // Lobby
	RoundIntro      GameState = 1
	ShowQuestion    GameState = 2
	ShowAnswers     GameState = 3
	RevealTheTruth  GameState = 4
	ScoreBoard      GameState = 5
	ScoreBoardFinal GameState = 6
)

func (s GameState) String() string {
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
func (s GameState) Duration() int {
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
