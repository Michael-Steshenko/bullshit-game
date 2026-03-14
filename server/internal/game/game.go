package game

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MaxPlayers   = 8
	MaxNickLen   = 9
	MaxAnswerLen = 40
)

// Player represents a player in a game.
type Player struct {
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
	Index    int    `json:"index"`
}

// Answer represents a submitted fake answer.
type Answer struct {
	PlayerUUID string
	Text       string
}

// Selection represents a player's vote.
type Selection struct {
	PlayerUUID string
	AnswerText string
}

// RevealAnswer represents one answer in the reveal sequence.
type RevealAnswer struct {
	Text       string   `json:"text"`
	Creators   []string `json:"creators"`  // UUIDs of players who wrote this (or "house" / "truth")
	Selectors  []string `json:"selectors"` // UUIDs of players who selected this
	RealAnswer bool     `json:"realAnswer"`
	HouseLie   bool     `json:"houseLie"`
	Points     int      `json:"points"` // Points earned/lost per selector
}

// Game holds the complete state of a single game.
type Game struct {
	StateTimestamp  time.Time
	CurrentQuestion *Question
	Selections      map[string]*Selection
	Answers         map[string]*Answer
	Players         map[string]*Player
	HostID          string
	Lang            string
	PIN             string
	PlayerOrder     []string
	RevealAnswers   []RevealAnswer
	QuestionIDs     []int
	Questions       []Question
	QuestionIndex   int
	TotalQuestions  int
	RoundIndex      int
	StateVersion    int
	State           State
	mu              sync.RWMutex
}

// NewGame creates a new game in staging state.
func NewGame(pin, hostUUID, hostNickname, lang string, totalQuestions int, questions []Question) *Game {
	g := &Game{
		PIN:            pin,
		State:          GameStaging,
		StateTimestamp: time.Now(),
		StateVersion:   0,
		Lang:           lang,
		TotalQuestions: totalQuestions,
		Players:        make(map[string]*Player),
		PlayerOrder:    make([]string, 0),
		Answers:        make(map[string]*Answer),
		Selections:     make(map[string]*Selection),
		Questions:      questions,
	}

	// Add the host as the first player
	g.addPlayerLocked(hostUUID, hostNickname)
	g.HostID = hostUUID

	return g
}

// AddPlayer adds a player to the game. Returns error string or empty.
func (g *Game) AddPlayer(uuid, nickname string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != GameStaging {
		return "GAME_STARTED"
	}
	if len(g.Players) >= MaxPlayers {
		return "GAME_IS_FULL"
	}

	g.addPlayerLocked(uuid, nickname)

	// First player becomes the host
	if g.HostID == "" {
		g.HostID = uuid
	}

	return ""
}

func (g *Game) addPlayerLocked(uuid, nickname string) {
	if len(nickname) > MaxNickLen {
		nickname = nickname[:MaxNickLen]
	}
	g.Players[uuid] = &Player{
		UUID:     uuid,
		Nickname: nickname,
		Score:    0,
		Index:    len(g.PlayerOrder),
	}
	g.PlayerOrder = append(g.PlayerOrder, uuid)
}

// GetPlayer returns a player by UUID, or nil.
func (g *Game) GetPlayer(uuid string) *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Players[uuid]
}

// GetPlayers returns all players in join order.
func (g *Game) GetPlayers() []*Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]*Player, 0, len(g.PlayerOrder))
	for _, uid := range g.PlayerOrder {
		if p, ok := g.Players[uid]; ok {
			result = append(result, p)
		}
	}
	return result
}

// PlayerCount returns the number of players.
func (g *Game) PlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Players)
}

// IsHost returns true if the given UUID is the host.
func (g *Game) IsHost(uuid string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.HostID == uuid
}

// StartGame transitions from staging to round intro.
func (g *Game) StartGame(uuid string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.HostID != uuid || g.State != GameStaging {
		return false
	}

	g.QuestionIndex = 0
	g.RoundIndex = 0
	g.loadCurrentQuestion()
	g.advanceStateLocked(RoundIntro)
	return true
}

// SubmitAnswer handles a player's answer submission.
// Returns error code or empty string on success.
func (g *Game) SubmitAnswer(uuid, text string, stateVersion int) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != ShowQuestion || g.StateVersion != stateVersion {
		return "INVALID_STATE"
	}
	if _, ok := g.Players[uuid]; !ok {
		return "PLAYER_NOT_FOUND"
	}

	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return "EMPTY_ANSWER"
	}
	if len(normalized) > MaxAnswerLen {
		return "ANSWER_TOO_LONG"
	}
	if g.CurrentQuestion != nil && strings.ToLower(strings.TrimSpace(g.CurrentQuestion.RealAnswer)) == normalized {
		return "CORRECT_ANSWER"
	}

	g.Answers[uuid] = &Answer{
		PlayerUUID: uuid,
		Text:       normalized,
	}
	return ""
}

// AllAnswersSubmitted returns true if all players have submitted.
func (g *Game) AllAnswersSubmitted() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Answers) >= len(g.Players)
}

// SelectAnswer handles a player's vote.
func (g *Game) SelectAnswer(uuid, text string, stateVersion int) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != ShowAnswers || g.StateVersion != stateVersion {
		return "INVALID_STATE"
	}
	if _, ok := g.Players[uuid]; !ok {
		return "PLAYER_NOT_FOUND"
	}

	g.Selections[uuid] = &Selection{
		PlayerUUID: uuid,
		AnswerText: strings.ToLower(strings.TrimSpace(text)),
	}
	return ""
}

// AllSelectionsSubmitted returns true if all players have voted.
func (g *Game) AllSelectionsSubmitted() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Selections) >= len(g.Players)
}

// Tick handles a timer expiration. Returns true if state was advanced.
func (g *Game) Tick(stateVersion int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.StateVersion != stateVersion {
		return false
	}

	return g.advanceToNextState()
}

// advanceToNextState moves the game to the next logical state.
// Must be called with lock held.
func (g *Game) advanceToNextState() bool {
	switch g.State {
	case RoundIntro:
		g.advanceStateLocked(ShowQuestion)
		return true
	case ShowQuestion:
		g.buildAnswerPool()
		g.advanceStateLocked(ShowAnswers)
		return true
	case ShowAnswers:
		g.computeReveal()
		g.computeScores()
		g.advanceStateLocked(RevealTheTruth)
		return true
	case RevealTheTruth:
		if g.isLastQuestion() {
			g.advanceStateLocked(ScoreBoardFinal)
		} else {
			g.advanceStateLocked(ScoreBoard)
		}
		return true
	case ScoreBoard:
		g.QuestionIndex++
		g.RoundIndex = g.computeRoundIndex()
		g.loadCurrentQuestion()
		g.clearRoundData()
		g.advanceStateLocked(RoundIntro)
		return true
	default:
		return false
	}
}

func (g *Game) advanceStateLocked(newState State) {
	g.State = newState
	g.StateTimestamp = time.Now()
	g.StateVersion++
}

// computeRoundIndex determines the round based on question index.
func (g *Game) computeRoundIndex() int {
	if g.TotalQuestions <= 1 {
		return 2
	}
	half := g.TotalQuestions / 2
	if g.QuestionIndex < half {
		return 0
	}
	if g.QuestionIndex < g.TotalQuestions-1 {
		return 1
	}
	return 2
}

func (g *Game) isLastQuestion() bool {
	return g.QuestionIndex >= g.TotalQuestions-1
}

func (g *Game) loadCurrentQuestion() {
	if g.QuestionIndex < len(g.Questions) {
		q := g.Questions[g.QuestionIndex]
		g.CurrentQuestion = &q
	}
}

func (g *Game) clearRoundData() {
	g.Answers = make(map[string]*Answer)
	g.Selections = make(map[string]*Selection)
	g.RevealAnswers = nil
}

// buildAnswerPool assembles answers with house lies for the ShowAnswers phase.
// No return value; the pool is used internally in computeReveal.
func (g *Game) buildAnswerPool() {
	// This is a no-op here; the answer pool is computed during reveal.
}

// GetAnswersForPlayer returns the shuffled answer list excluding the player's own.
func (g *Game) GetAnswersForPlayer(uuid string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	answerSet := g.buildFullAnswerSet()

	// Exclude the player's own answer
	myAnswer := ""
	if a, ok := g.Answers[uuid]; ok {
		myAnswer = a.Text
	}

	var result []string
	myAnswerExcluded := false
	for _, a := range answerSet {
		if a == myAnswer && !myAnswerExcluded {
			myAnswerExcluded = true
			continue
		}
		result = append(result, a)
	}
	return result
}

// buildFullAnswerSet creates the complete answer pool: unique player answers + house lies + real answer.
func (g *Game) buildFullAnswerSet() []string {
	uniqueAnswers := make(map[string]bool)
	for _, a := range g.Answers {
		uniqueAnswers[a.Text] = true
	}

	playerCount := len(g.Players)
	uniqueCount := len(uniqueAnswers)
	houseLiesNeeded := playerCount - uniqueCount

	var answers []string
	for text := range uniqueAnswers {
		answers = append(answers, text)
	}

	// Add house lies
	if g.CurrentQuestion != nil && houseLiesNeeded > 0 {
		for i := 0; i < houseLiesNeeded && i < len(g.CurrentQuestion.FakeAnswers); i++ {
			lie := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.FakeAnswers[i]))
			if !uniqueAnswers[lie] {
				answers = append(answers, lie)
				uniqueAnswers[lie] = true
			}
		}
	}

	// Add real answer
	if g.CurrentQuestion != nil {
		realAnswer := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.RealAnswer))
		if !uniqueAnswers[realAnswer] {
			answers = append(answers, realAnswer)
		}
	}

	sort.Strings(answers)
	return answers
}

// computeReveal builds the reveal sequence.
func (g *Game) computeReveal() {
	if g.CurrentQuestion == nil {
		return
	}

	realAnswer := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.RealAnswer))

	// Collect unique player answers and their creators
	playerAnswerCreators := make(map[string][]string) // answer text -> []creator UUIDs
	for uuid, a := range g.Answers {
		playerAnswerCreators[a.Text] = append(playerAnswerCreators[a.Text], uuid)
	}

	// Determine house lies needed
	uniquePlayerAnswers := make(map[string]bool)
	for _, a := range g.Answers {
		uniquePlayerAnswers[a.Text] = true
	}
	houseLiesNeeded := len(g.Players) - len(uniquePlayerAnswers)

	houseLies := make(map[string]bool)
	if houseLiesNeeded > 0 {
		for i := 0; i < houseLiesNeeded && i < len(g.CurrentQuestion.FakeAnswers); i++ {
			lie := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.FakeAnswers[i]))
			if !uniquePlayerAnswers[lie] && lie != realAnswer {
				houseLies[lie] = true
			}
		}
	}

	// Collect selectors per answer
	answerSelectors := make(map[string][]string) // answer text -> []selector UUIDs
	for uuid, sel := range g.Selections {
		answerSelectors[sel.AnswerText] = append(answerSelectors[sel.AnswerText], uuid)
	}

	// Build reveal entries
	var reveals []RevealAnswer

	// Player lies
	for text, creators := range playerAnswerCreators {
		if text == realAnswer {
			continue
		}
		selectors := answerSelectors[text]
		points := 0
		if len(selectors) > 0 {
			// Selecting another player's lie: 0 points for selector, creator gets bullshit points
			points = 0
		}
		reveals = append(reveals, RevealAnswer{
			Text:       text,
			Creators:   creators,
			Selectors:  selectors,
			RealAnswer: false,
			HouseLie:   false,
			Points:     points,
		})
	}

	// House lies
	for text := range houseLies {
		selectors := answerSelectors[text]
		reveals = append(reveals, RevealAnswer{
			Text:       text,
			Creators:   []string{"house"},
			Selectors:  selectors,
			RealAnswer: false,
			HouseLie:   true,
			Points:     ScoreForHouseLiePenalty(g.RoundIndex),
		})
	}

	// Sort alphabetically, but real answer always last
	sort.Slice(reveals, func(i, j int) bool {
		return reveals[i].Text < reveals[j].Text
	})

	// Real answer (always last)
	realSelectors := answerSelectors[realAnswer]
	reveals = append(reveals, RevealAnswer{
		Text:       realAnswer,
		Creators:   []string{"truth"},
		Selectors:  realSelectors,
		RealAnswer: true,
		HouseLie:   false,
		Points:     ScoreForCorrectAnswer(g.RoundIndex),
	})

	g.RevealAnswers = reveals
}

// computeScores updates player scores based on current round's answers and selections.
func (g *Game) computeScores() {
	if g.CurrentQuestion == nil {
		return
	}

	realAnswer := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.RealAnswer))

	// Collect unique player answers for house lie detection
	uniquePlayerAnswers := make(map[string]bool)
	for _, a := range g.Answers {
		uniquePlayerAnswers[a.Text] = true
	}

	houseLiesNeeded := len(g.Players) - len(uniquePlayerAnswers)
	houseLies := make(map[string]bool)
	if houseLiesNeeded > 0 {
		for i := 0; i < houseLiesNeeded && i < len(g.CurrentQuestion.FakeAnswers); i++ {
			lie := strings.ToLower(strings.TrimSpace(g.CurrentQuestion.FakeAnswers[i]))
			if !uniquePlayerAnswers[lie] && lie != realAnswer {
				houseLies[lie] = true
			}
		}
	}

	// Map answer text to creator UUIDs
	answerCreators := make(map[string][]string)
	for uuid, a := range g.Answers {
		answerCreators[a.Text] = append(answerCreators[a.Text], uuid)
	}

	// Process each selection
	for uuid, sel := range g.Selections {
		player := g.Players[uuid]
		if player == nil {
			continue
		}

		selectedText := sel.AnswerText

		if selectedText == realAnswer {
			// Selected the truth
			player.Score += ScoreForCorrectAnswer(g.RoundIndex)
		} else if houseLies[selectedText] {
			// Selected a house lie
			player.Score += ScoreForHouseLiePenalty(g.RoundIndex)
		}
		// Selecting another player's lie: 0 points for the selector
	}

	// Award bullshit points to answer creators who fooled players
	for text, creators := range answerCreators {
		if text == realAnswer {
			continue
		}
		// Count how many players selected this answer
		fooledCount := 0
		for _, sel := range g.Selections {
			if sel.AnswerText == text {
				fooledCount++
			}
		}
		if fooledCount > 0 {
			pointsPerCreator := ScoreForFoolingPlayer(g.RoundIndex) * fooledCount / len(creators)
			for _, creatorUUID := range creators {
				if p := g.Players[creatorUUID]; p != nil {
					p.Score += pointsPerCreator
				}
			}
		}
	}
}

// Rematch resets the game for a new match. Same PIN, same players, new questions.
func (g *Game) Rematch(questions []Question) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Reset scores
	for _, p := range g.Players {
		p.Score = 0
	}

	g.State = GameStaging
	g.StateTimestamp = time.Now()
	g.StateVersion = 0
	g.RoundIndex = 0
	g.QuestionIndex = 0
	g.Questions = questions
	g.TotalQuestions = len(questions)
	g.CurrentQuestion = nil
	g.clearRoundData()
}

// GetRevealAnswers returns a copy of the reveal answers.
func (g *Game) GetRevealAnswers() []RevealAnswer {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]RevealAnswer, len(g.RevealAnswers))
	copy(result, g.RevealAnswers)
	return result
}

// GetCurrentQuestion returns the current question (thread-safe).
func (g *Game) GetCurrentQuestion() *Question {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.CurrentQuestion == nil {
		return nil
	}
	q := *g.CurrentQuestion
	return &q
}

// GetStateSnapshot returns a snapshot of the current game state.
func (g *Game) GetStateSnapshot() StateSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return StateSnapshot{
		State:          g.State,
		StateTimestamp: g.StateTimestamp,
		StateVersion:   g.StateVersion,
		RoundIndex:     g.RoundIndex,
		QuestionIndex:  g.QuestionIndex,
		TotalQuestions: g.TotalQuestions,
	}
}

// StateSnapshot is a read-only snapshot of game state for broadcasting.
type StateSnapshot struct {
	StateTimestamp time.Time `json:"stateTimestamp"`
	State          State     `json:"state"`
	StateVersion   int       `json:"stateVersion"`
	RoundIndex     int       `json:"roundIndex"`
	QuestionIndex  int       `json:"questionIndex"`
	TotalQuestions int       `json:"totalQuestions"`
}
