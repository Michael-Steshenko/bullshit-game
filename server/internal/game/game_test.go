package game

import (
	"testing"
)

func makeTestQuestions(n int) []Question {
	questions := make([]Question, n)
	for i := 0; i < n; i++ {
		questions[i] = Question{
			ID:          i + 1,
			Lang:        "en",
			Text:        "What is $blank$?",
			RealAnswer:  "the truth",
			FakeAnswers: []string{"house lie 1", "house lie 2", "house lie 3"},
			Citation:    "test",
		}
	}
	return questions
}

func TestNewGame(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host-uuid", "Host", "en", 7, questions)

	if g.PIN != "ABCD" {
		t.Errorf("PIN = %s, want ABCD", g.PIN)
	}
	if g.State != GameStaging {
		t.Errorf("State = %v, want GameStaging", g.State)
	}
	if g.PlayerCount() != 1 {
		t.Errorf("PlayerCount = %d, want 1", g.PlayerCount())
	}
	if !g.IsHost("host-uuid") {
		t.Error("host-uuid should be host")
	}
}

func TestAddPlayer(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)

	err := g.AddPlayer("p1", "Player1")
	if err != "" {
		t.Errorf("AddPlayer returned error: %s", err)
	}

	if g.PlayerCount() != 2 {
		t.Errorf("PlayerCount = %d, want 2", g.PlayerCount())
	}
}

func TestAddPlayerMaxPlayers(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)

	for i := 0; i < 7; i++ {
		g.AddPlayer("p"+string(rune('1'+i)), "P")
	}

	err := g.AddPlayer("extra", "Extra")
	if err != "GAME_IS_FULL" {
		t.Errorf("expected GAME_IS_FULL, got %s", err)
	}
}

func TestAddPlayerAfterStart(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.StartGame("host")

	err := g.AddPlayer("late", "Late")
	if err != "GAME_STARTED" {
		t.Errorf("expected GAME_STARTED, got %s", err)
	}
}

func TestStartGame(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)

	ok := g.StartGame("host")
	if !ok {
		t.Error("StartGame should succeed for host")
	}
	if g.State != RoundIntro {
		t.Errorf("State = %v, want RoundIntro", g.State)
	}
}

func TestStartGameNonHost(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.AddPlayer("p1", "Player1")

	ok := g.StartGame("p1")
	if ok {
		t.Error("StartGame should fail for non-host")
	}
}

func TestSubmitAnswer(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.AddPlayer("p1", "Player1")
	g.StartGame("host")
	// Advance to ShowQuestion
	g.Tick(g.StateVersion)

	sv := g.StateVersion
	err := g.SubmitAnswer("host", "my lie", sv)
	if err != "" {
		t.Errorf("SubmitAnswer returned: %s", err)
	}
}

func TestSubmitAnswerEmpty(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.StartGame("host")
	g.Tick(g.StateVersion)

	sv := g.StateVersion
	err := g.SubmitAnswer("host", "  ", sv)
	if err != "EMPTY_ANSWER" {
		t.Errorf("expected EMPTY_ANSWER, got %s", err)
	}
}

func TestSubmitAnswerTooLong(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.StartGame("host")
	g.Tick(g.StateVersion)

	sv := g.StateVersion
	longAnswer := "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddx" // 41 chars
	err := g.SubmitAnswer("host", longAnswer, sv)
	if err != "ANSWER_TOO_LONG" {
		t.Errorf("expected ANSWER_TOO_LONG, got %s", err)
	}
}

func TestSubmitCorrectAnswer(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.StartGame("host")
	g.Tick(g.StateVersion)

	sv := g.StateVersion
	err := g.SubmitAnswer("host", "the truth", sv)
	if err != "CORRECT_ANSWER" {
		t.Errorf("expected CORRECT_ANSWER, got %s", err)
	}
}

func TestSubmitAnswerStaleVersion(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)
	g.StartGame("host")
	g.Tick(g.StateVersion)

	err := g.SubmitAnswer("host", "my lie", 0)
	if err != "INVALID_STATE" {
		t.Errorf("expected INVALID_STATE, got %s", err)
	}
}

func TestComputeRoundIndex(t *testing.T) {
	questions := makeTestQuestions(7)
	g := NewGame("ABCD", "host", "Host", "en", 7, questions)

	tests := []struct {
		qIndex   int
		expected int
	}{
		{0, 0}, {1, 0}, {2, 0}, // Round 0: 0-2
		{3, 1}, {4, 1}, {5, 1}, // Round 1: 3-5
		{6, 2}, // Round 2: 6
	}

	for _, tt := range tests {
		g.QuestionIndex = tt.qIndex
		got := g.computeRoundIndex()
		if got != tt.expected {
			t.Errorf("questionIndex=%d: roundIndex = %d, want %d", tt.qIndex, got, tt.expected)
		}
	}
}

func TestFullGameFlow(t *testing.T) {
	questions := makeTestQuestions(1)
	g := NewGame("ABCD", "host", "Host", "en", 1, questions)
	g.AddPlayer("p1", "Player1")

	// Start -> RoundIntro
	g.StartGame("host")
	if g.State != RoundIntro {
		t.Fatalf("expected RoundIntro, got %v", g.State)
	}

	// RoundIntro -> ShowQuestion
	g.Tick(g.StateVersion)
	if g.State != ShowQuestion {
		t.Fatalf("expected ShowQuestion, got %v", g.State)
	}

	// Submit answers
	sv := g.StateVersion
	g.SubmitAnswer("host", "fake answer 1", sv)
	g.SubmitAnswer("p1", "fake answer 2", sv)

	// ShowQuestion -> ShowAnswers (auto via AllAnswersSubmitted)
	g.Tick(g.StateVersion)
	if g.State != ShowAnswers {
		t.Fatalf("expected ShowAnswers, got %v", g.State)
	}

	// Select answers
	sv = g.StateVersion
	g.SelectAnswer("host", "the truth", sv)
	g.SelectAnswer("p1", "fake answer 1", sv)

	// ShowAnswers -> RevealTheTruth
	g.Tick(g.StateVersion)
	if g.State != RevealTheTruth {
		t.Fatalf("expected RevealTheTruth, got %v", g.State)
	}

	// RevealTheTruth -> ScoreBoardFinal (last question)
	g.Tick(g.StateVersion)
	if g.State != ScoreBoardFinal {
		t.Fatalf("expected ScoreBoardFinal, got %v", g.State)
	}

	// Check scores
	host := g.GetPlayer("host")
	p1 := g.GetPlayer("p1")

	// Host: +1000 (correct) + 500 (fooled p1)
	if host.Score != 1500 {
		t.Errorf("host score = %d, want 1500", host.Score)
	}

	// P1: 0 (selected player lie, no penalty, no correct)
	if p1.Score != 0 {
		t.Errorf("p1 score = %d, want 0", p1.Score)
	}
}

func TestHouseLiePenalty(t *testing.T) {
	questions := []Question{{
		ID:          1,
		Lang:        "en",
		Text:        "What is $blank$?",
		RealAnswer:  "the truth",
		FakeAnswers: []string{"house lie 1", "house lie 2"},
		Citation:    "test",
	}}
	g := NewGame("ABCD", "host", "Host", "en", 1, questions)
	g.AddPlayer("p1", "Player1")

	g.StartGame("host")
	g.Tick(g.StateVersion) // -> ShowQuestion

	sv := g.StateVersion
	g.SubmitAnswer("host", "my lie", sv)
	// p1 doesn't submit -> house lie fills slot

	g.Tick(g.StateVersion) // -> ShowAnswers

	// p1 selects a house lie
	sv = g.StateVersion
	g.SelectAnswer("host", "the truth", sv)
	g.SelectAnswer("p1", "house lie 1", sv)

	g.Tick(g.StateVersion) // -> RevealTheTruth

	p1 := g.GetPlayer("p1")
	// p1 selected house lie in round 0: -500
	if p1.Score != -500 {
		t.Errorf("p1 score = %d, want -500", p1.Score)
	}
}

func TestRematch(t *testing.T) {
	questions := makeTestQuestions(1)
	g := NewGame("ABCD", "host", "Host", "en", 1, questions)
	g.AddPlayer("p1", "Player1")
	g.StartGame("host")

	// Play through
	g.Tick(g.StateVersion) // ShowQuestion
	sv := g.StateVersion
	g.SubmitAnswer("host", "lie", sv)
	g.SubmitAnswer("p1", "lie2", sv)
	g.Tick(g.StateVersion) // ShowAnswers
	sv = g.StateVersion
	g.SelectAnswer("host", "the truth", sv)
	g.SelectAnswer("p1", "the truth", sv)
	g.Tick(g.StateVersion) // RevealTheTruth
	g.Tick(g.StateVersion) // ScoreBoardFinal

	// Rematch
	newQ := makeTestQuestions(1)
	g.Rematch(newQ)

	if g.State != GameStaging {
		t.Errorf("State = %v, want GameStaging", g.State)
	}
	if g.StateVersion != 0 {
		t.Errorf("StateVersion = %d, want 0", g.StateVersion)
	}
	if g.GetPlayer("host").Score != 0 {
		t.Errorf("host score not reset")
	}
	if g.PlayerCount() != 2 {
		t.Errorf("PlayerCount = %d, want 2", g.PlayerCount())
	}
}

func TestRevealOmitsUnselectedAnswers(t *testing.T) {
	questions := makeTestQuestions(1)
	g := NewGame("ABCD", "host", "Host", "en", 1, questions)
	g.AddPlayer("p1", "Player1")
	g.AddPlayer("p2", "Player2")

	g.StartGame("host")
	g.Tick(g.StateVersion) // -> ShowQuestion

	sv := g.StateVersion
	if err := g.SubmitAnswer("host", "host lie", sv); err != "" {
		t.Fatalf("submit host: %s", err)
	}
	if err := g.SubmitAnswer("p1", "p1 lie", sv); err != "" {
		t.Fatalf("submit p1: %s", err)
	}
	if err := g.SubmitAnswer("p2", "p2 lie", sv); err != "" {
		t.Fatalf("submit p2: %s", err)
	}

	g.Tick(g.StateVersion) // -> ShowAnswers

	sv = g.StateVersion
	if err := g.SelectAnswer("host", "the truth", sv); err != "" {
		t.Fatalf("select host: %s", err)
	}
	if err := g.SelectAnswer("p1", "host lie", sv); err != "" {
		t.Fatalf("select p1: %s", err)
	}
	if err := g.SelectAnswer("p2", "the truth", sv); err != "" {
		t.Fatalf("select p2: %s", err)
	}

	g.Tick(g.StateVersion) // -> RevealTheTruth

	reveals := g.GetRevealAnswers()
	if len(reveals) != 2 {
		t.Fatalf("reveal count = %d, want 2", len(reveals))
	}

	var hasHostLie, hasTruth, hasP1Lie, hasP2Lie bool
	for _, r := range reveals {
		switch r.Text {
		case "host lie":
			hasHostLie = true
		case "the truth":
			hasTruth = true
		case "p1 lie":
			hasP1Lie = true
		case "p2 lie":
			hasP2Lie = true
		}
	}

	if !hasHostLie || !hasTruth {
		t.Fatalf("expected reveal to include host lie and truth, got %+v", reveals)
	}
	if hasP1Lie || hasP2Lie {
		t.Fatalf("unexpected unselected answer in reveal: %+v", reveals)
	}
}

func TestRevealIncludesCreatorPointsWhenBullshitting(t *testing.T) {
	questions := []Question{{
		ID:          1,
		Lang:        "en",
		Text:        "What is $blank$?",
		RealAnswer:  "the truth",
		FakeAnswers: []string{"house lie 1", "house lie 2"},
		Citation:    "test",
	}}
	g := NewGame("ABCD", "host", "Host", "en", 1, questions)
	g.AddPlayer("p1", "Player1")

	g.StartGame("host")
	g.Tick(g.StateVersion) // -> ShowQuestion

	sv := g.StateVersion
	if err := g.SubmitAnswer("host", "host lie", sv); err != "" {
		t.Fatalf("submit host: %s", err)
	}
	// p1 doesn't submit -> house lie added

	g.Tick(g.StateVersion) // -> ShowAnswers

	sv = g.StateVersion
	if err := g.SelectAnswer("host", "the truth", sv); err != "" {
		t.Fatalf("select host: %s", err)
	}
	if err := g.SelectAnswer("p1", "host lie", sv); err != "" {
		t.Fatalf("select p1: %s", err)
	}

	g.Tick(g.StateVersion) // -> RevealTheTruth

	reveals := g.GetRevealAnswers()
	var hostLie *RevealAnswer
	for i := range reveals {
		if reveals[i].Text == "host lie" {
			hostLie = &reveals[i]
			break
		}
	}
	if hostLie == nil {
		t.Fatalf("host lie not found in reveal: %+v", reveals)
	}

	if hostLie.CreatorPoints != 500 {
		t.Fatalf("creator points = %d, want 500", hostLie.CreatorPoints)
	}
	if hostLie.SelectorPoints != 0 {
		t.Fatalf("selector points = %d, want 0", hostLie.SelectorPoints)
	}
}
