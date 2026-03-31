package hub

import "encoding/json"

// Incoming message types (client -> server)
const (
	MsgJoin         = "join"
	MsgReconnect    = "reconnect"
	MsgStartGame    = "start_game"
	MsgSubmitAnswer = "submit_answer"
	MsgSelectAnswer = "select_answer"
	MsgTick         = "tick"
	MsgRematch      = "rematch"
)

// Outgoing message types (server -> client)
const (
	MsgGameState       = "game_state"
	MsgRejoined        = "rejoined"
	MsgPlayerJoined    = "player_joined"
	MsgPlayerList      = "player_list"
	MsgQuestion        = "question"
	MsgAnswerSubmitted = "answer_submitted"
	MsgAnswers         = "answers"
	MsgAnswerSelected  = "answer_selected"
	MsgReveal          = "reveal"
	MsgScores          = "scores"
	MsgFinalScores     = "final_scores"
	MsgRematchResp     = "rematch"
	MsgError           = "error"
	MsgTimeSync        = "time_sync"
)

// IncomingMessage is a generic incoming WebSocket message.
type IncomingMessage struct {
	Type         string `json:"type"`
	PIN          string `json:"pin,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	Text         string `json:"text,omitempty"`
	StateVersion int    `json:"stateVersion,omitempty"`
}

// OutgoingMessage is a generic outgoing WebSocket message.
type OutgoingMessage struct {
	Data interface{} `json:"data,omitempty"`
	Type string      `json:"type"`
}

func NewOutgoing(msgType string, data interface{}) []byte {
	msg := OutgoingMessage{Type: msgType, Data: data}
	b, _ := json.Marshal(msg)
	return b
}

// Error payloads
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// State payloads
type GameStatePayload struct {
	State          int   `json:"state"`
	StateTimestamp int64 `json:"stateTimestamp"`
	StateVersion   int   `json:"stateVersion"`
	RoundIndex     int   `json:"roundIndex"`
	QuestionIndex  int   `json:"questionIndex"`
	TotalQuestions int   `json:"totalQuestions"`
	Duration       int   `json:"duration"`
}

type PlayerPayload struct {
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
	Index    int    `json:"index"`
}

type PlayerListPayload struct {
	Players []PlayerPayload `json:"players"`
}

type PlayerJoinedPayload struct {
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
	Index    int    `json:"index"`
}

type RejoinedPayload struct {
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
	Index    int    `json:"index"`
}

type QuestionPayload struct {
	Text           string `json:"text"`
	QuestionNumber int    `json:"questionNumber"`
	TotalQuestions int    `json:"totalQuestions"`
}

type AnswerSubmittedPayload struct {
	UUID string `json:"uuid"`
}

type AnswersPayload struct {
	Answers []AnswerOption `json:"answers"`
}

type AnswerOption struct {
	Text string `json:"text"`
}

type AnswerSelectedPayload struct {
	UUID string `json:"uuid"`
}

type RevealPayload struct {
	Reveals []RevealEntry `json:"reveals"`
}

type RevealEntry struct {
	Text           string   `json:"text"`
	Selectors      []string `json:"selectors"`
	Creators       []string `json:"creators"`
	RealAnswer     bool     `json:"realAnswer"`
	HouseLie       bool     `json:"houseLie"`
	SelectorPoints int      `json:"selectorPoints"`
	CreatorPoints  int      `json:"creatorPoints"`
}

type ScoresPayload struct {
	Players []PlayerPayload `json:"players"`
}

type TimeSyncPayload struct {
	ServerTime int64 `json:"serverTime"`
}
