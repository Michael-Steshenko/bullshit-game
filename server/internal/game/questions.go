package game

import (
	"context"
	"database/sql"
	"math/rand"
)

// Question represents a trivia question.
type Question struct {
	ID          int      `json:"id"`
	Lang        string   `json:"lang"`
	Text        string   `json:"questionText"`
	RealAnswer  string   `json:"realAnswer"`
	FakeAnswers []string `json:"fakeAnswers"`
	Citation    string   `json:"citation"`
}

// QuestionStore is an interface for fetching questions.
type QuestionStore interface {
	// GetRandomQuestions returns n random questions for the given language.
	GetRandomQuestions(ctx context.Context, lang string, n int) ([]Question, error)
}

// DBQuestionStore fetches questions from Postgres.
type DBQuestionStore struct {
	db *sql.DB
}

func NewDBQuestionStore(db *sql.DB) *DBQuestionStore {
	return &DBQuestionStore{db: db}
}

func (s *DBQuestionStore) GetRandomQuestions(ctx context.Context, lang string, n int) ([]Question, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, lang, question, real_answer, fake_answers, COALESCE(citation, '')
		FROM questions
		WHERE lang = $1
		ORDER BY random()
		LIMIT $2
	`, lang, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		var fakeAnswers []byte
		if err := rows.Scan(&q.ID, &q.Lang, &q.Text, &q.RealAnswer, &fakeAnswers, &q.Citation); err != nil {
			return nil, err
		}
		q.FakeAnswers = parsePostgresArray(fakeAnswers)
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// parsePostgresArray parses a Postgres text[] into a Go string slice.
func parsePostgresArray(b []byte) []string {
	s := string(b)
	if s == "{}" || s == "" {
		return nil
	}
	// Remove outer braces
	s = s[1 : len(s)-1]

	var result []string
	var current []byte
	inQuote := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			current = append(current, ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ',' && !inQuote {
			result = append(result, string(current))
			current = current[:0]
			continue
		}
		current = append(current, ch)
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

// ShuffleQuestionIDs returns a shuffled copy of question IDs.
func ShuffleQuestionIDs(ids []int) []int {
	shuffled := make([]int, len(ids))
	copy(shuffled, ids)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}
