package hub

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bullshit-wtf/server/internal/game"
	"github.com/google/uuid"
)

// Hub manages all game rooms and WebSocket connections.
type Hub struct {
	questionStore game.QuestionStore
	rooms         map[string]*Room
	pinGen        *game.PinGenerator
	register      chan *Client
	unregister    chan *Client
	mu            sync.RWMutex
}

func NewHub(qs game.QuestionStore, pg *game.PinGenerator) *Hub {
	return &Hub{
		rooms:         make(map[string]*Room),
		questionStore: qs,
		pinGen:        pg,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Registration happens in processMessage
			_ = client

		case client := <-h.unregister:
			if client.PIN != "" && client.UUID != "" {
				h.mu.RLock()
				room := h.rooms[client.PIN]
				h.mu.RUnlock()
				if room != nil {
					room.RemoveClient(client.UUID)
				}
			}
		}
	}
}

// GetRoom returns the room for a given PIN.
func (h *Hub) GetRoom(pin string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[strings.ToUpper(pin)]
}

// processMessage handles an incoming WebSocket message.
func (h *Hub) processMessage(client *Client, msg *IncomingMessage) {
	switch msg.Type {
	case MsgCreateAndJoin:
		h.handleCreateAndJoin(client, msg)
	case MsgValidatePIN:
		h.handleValidatePIN(client, msg)
	case MsgJoin:
		h.handleJoin(client, msg)
	case MsgReconnect:
		h.handleReconnect(client, msg)
	case MsgStartGame:
		h.handleStartGame(client, msg)
	case MsgSubmitAnswer:
		h.handleSubmitAnswer(client, msg)
	case MsgSelectAnswer:
		h.handleSelectAnswer(client, msg)
	case MsgTick:
		h.handleTick(client, msg)
	case MsgRematch:
		h.handleRematch(client, msg)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (h *Hub) handleCreateAndJoin(client *Client, msg *IncomingMessage) {
	nickname := strings.TrimSpace(msg.Nickname)
	if nickname == "" {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "EMPTY_NICKNAME", Message: "Nickname required"}))
		return
	}

	lang := strings.TrimSpace(msg.Lang)
	if lang == "" {
		lang = "en"
	}
	totalQuestions := msg.TotalQuestions
	if totalQuestions <= 0 {
		totalQuestions = 7
	}

	pin := h.pinGen.Next()
	questions, err := h.questionStore.GetRandomQuestions(context.Background(), lang, totalQuestions)
	if err != nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "CREATE_GAME_FAILED", Message: "Failed to create game"}))
		return
	}

	playerUUID := uuid.New().String()
	g := game.NewGame(pin, playerUUID, nickname, lang, len(questions), questions)
	room := NewRoom(g)

	h.mu.Lock()
	h.rooms[pin] = room
	h.mu.Unlock()

	client.UUID = playerUUID
	client.PIN = pin
	room.AddClient(playerUUID, client)

	player := room.Game.GetPlayer(playerUUID)
	if player == nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "PLAYER_NOT_FOUND", Message: "Player not found"}))
		return
	}

	client.Send(NewOutgoing(MsgCreatedGame, CreatedGamePayload{PIN: pin}))
	client.Send(NewOutgoing(MsgRejoined, RejoinedPayload{
		UUID:     playerUUID,
		Nickname: player.Nickname,
		Score:    player.Score,
		Index:    player.Index,
	}))
	client.Send(NewOutgoing(MsgPlayerList, h.buildPlayerList(room)))
	client.Send(h.buildGameStateMsg(room.Game))
}

func (h *Hub) handleValidatePIN(client *Client, msg *IncomingMessage) {
	pin := strings.ToUpper(msg.PIN)
	room := h.GetRoom(pin)
	if room == nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "GAME_NOT_EXIST", Message: "Wrong PIN"}))
		return
	}

	if room.Game.GetStateSnapshot().State != game.GameStaging {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "GAME_NOT_EXIST", Message: "Wrong PIN"}))
		return
	}

	client.Send(NewOutgoing(MsgPINValidated, PinValidatedPayload{PIN: pin}))
}

func (h *Hub) handleJoin(client *Client, msg *IncomingMessage) {
	pin := strings.ToUpper(msg.PIN)
	room := h.GetRoom(pin)
	if room == nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "GAME_NOT_EXIST", Message: "Wrong PIN"}))
		return
	}
	if room.Game.GetStateSnapshot().State != game.GameStaging {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "GAME_NOT_EXIST", Message: "Wrong PIN"}))
		return
	}

	playerUUID := uuid.New().String()
	nickname := msg.Nickname
	if len(nickname) > game.MaxNickLen {
		nickname = nickname[:game.MaxNickLen]
	}

	errCode := room.Game.AddPlayer(playerUUID, nickname)
	if errCode != "" {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: errCode, Message: errCode}))
		return
	}

	client.UUID = playerUUID
	client.PIN = pin
	room.AddClient(playerUUID, client)
	log.Printf("join: player %s (%s) joined pin=%s, total=%d, host=%s", playerUUID, nickname, pin, room.Game.PlayerCount(), room.Game.HostID)

	player := room.Game.GetPlayer(playerUUID)

	// Send rejoined to the new player (contains their UUID)
	client.Send(NewOutgoing(MsgRejoined, RejoinedPayload{
		UUID:     playerUUID,
		Nickname: player.Nickname,
		Score:    player.Score,
		Index:    player.Index,
	}))

	// Send player list to the new player
	client.Send(NewOutgoing(MsgPlayerList, h.buildPlayerList(room)))

	// Send game state
	client.Send(h.buildGameStateMsg(room.Game))

	// Broadcast player_joined to everyone else
	room.BroadcastExcept(NewOutgoing(MsgPlayerJoined, PlayerJoinedPayload{
		UUID:     playerUUID,
		Nickname: player.Nickname,
		Index:    player.Index,
	}), playerUUID)
}

func (h *Hub) handleReconnect(client *Client, msg *IncomingMessage) {
	pin := strings.ToUpper(msg.PIN)
	room := h.GetRoom(pin)
	if room == nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "RECONNECT_FAILED", Message: "Wrong PIN"}))
		return
	}

	player := room.Game.GetPlayer(msg.UUID)
	if player == nil {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: "RECONNECT_FAILED", Message: "Player not found"}))
		return
	}

	client.UUID = msg.UUID
	client.PIN = pin
	room.AddClient(msg.UUID, client)

	// Send rejoined
	client.Send(NewOutgoing(MsgRejoined, RejoinedPayload{
		UUID:     player.UUID,
		Nickname: player.Nickname,
		Score:    player.Score,
		Index:    player.Index,
	}))

	// Send player list
	client.Send(NewOutgoing(MsgPlayerList, h.buildPlayerList(room)))

	// Send game state
	client.Send(h.buildGameStateMsg(room.Game))

	// Send current state data
	h.sendCurrentStateData(client, room)
}

func (h *Hub) handleStartGame(client *Client, msg *IncomingMessage) {
	pin := strings.ToUpper(msg.PIN)
	room := h.GetRoom(pin)
	if room == nil {
		log.Printf("start_game: room not found for pin=%s", pin)
		return
	}

	log.Printf("start_game: pin=%s client.UUID=%s hostID=%s state=%v", pin, client.UUID, room.Game.HostID, room.Game.State)
	if !room.Game.StartGame(client.UUID) {
		log.Printf("start_game: StartGame returned false")
		return
	}

	// Broadcast game state (RoundIntro)
	room.Broadcast(h.buildGameStateMsg(room.Game))
}

func (h *Hub) handleSubmitAnswer(client *Client, msg *IncomingMessage) {
	room := h.GetRoom(client.PIN)
	if room == nil {
		return
	}

	errCode := room.Game.SubmitAnswer(client.UUID, msg.Text, msg.StateVersion)
	if errCode != "" {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: errCode, Message: errCode}))
		return
	}

	// Broadcast answer_submitted to all
	room.Broadcast(NewOutgoing(MsgAnswerSubmitted, AnswerSubmittedPayload{UUID: client.UUID}))

	// Check if all answers submitted -> auto advance
	if room.Game.AllAnswersSubmitted() {
		h.advanceGame(room)
	}
}

func (h *Hub) handleSelectAnswer(client *Client, msg *IncomingMessage) {
	room := h.GetRoom(client.PIN)
	if room == nil {
		return
	}

	errCode := room.Game.SelectAnswer(client.UUID, msg.Text, msg.StateVersion)
	if errCode != "" {
		client.Send(NewOutgoing(MsgError, ErrorPayload{Code: errCode, Message: errCode}))
		return
	}

	// Broadcast answer_selected to all
	room.Broadcast(NewOutgoing(MsgAnswerSelected, AnswerSelectedPayload{UUID: client.UUID}))

	// Check if all selections submitted -> auto advance
	if room.Game.AllSelectionsSubmitted() {
		h.advanceGame(room)
	}
}

func (h *Hub) handleTick(client *Client, msg *IncomingMessage) {
	room := h.GetRoom(client.PIN)
	if room == nil {
		return
	}

	if !room.Game.Tick(msg.StateVersion) {
		return
	}

	h.broadcastStateAndData(room)
}

func (h *Hub) handleRematch(client *Client, _ *IncomingMessage) {
	room := h.GetRoom(client.PIN)
	if room == nil {
		return
	}

	if !room.Game.IsHost(client.UUID) {
		return
	}

	questions, err := h.questionStore.GetRandomQuestions(context.Background(), room.Game.Lang, room.Game.TotalQuestions)
	if err != nil {
		log.Printf("rematch: failed to get questions: %v", err)
		return
	}

	room.Game.Rematch(questions)
	room.Broadcast(NewOutgoing(MsgRematchResp, nil))
	room.Broadcast(h.buildGameStateMsg(room.Game))
	room.Broadcast(NewOutgoing(MsgPlayerList, h.buildPlayerList(room)))
}

// advanceGame advances the game state and broadcasts to all.
func (h *Hub) advanceGame(room *Room) {
	snap := room.Game.GetStateSnapshot()
	if !room.Game.Tick(snap.StateVersion) {
		return
	}
	h.broadcastStateAndData(room)
}

func (h *Hub) broadcastStateAndData(room *Room) {
	// Broadcast game state
	room.Broadcast(h.buildGameStateMsg(room.Game))

	// Broadcast state-specific data
	snap := room.Game.GetStateSnapshot()
	switch game.State(snap.State) {
	case game.ShowQuestion:
		q := room.Game.GetCurrentQuestion()
		if q != nil {
			room.Broadcast(NewOutgoing(MsgQuestion, QuestionPayload{
				Text:           q.Text,
				QuestionNumber: snap.QuestionIndex + 1,
				TotalQuestions: snap.TotalQuestions,
			}))
		}
	case game.ShowAnswers:
		h.sendPersonalizedAnswers(room)
	case game.RevealTheTruth:
		revealAnswers := room.Game.GetRevealAnswers()
		reveals := make([]RevealEntry, len(revealAnswers))
		for i, r := range revealAnswers {
			reveals[i] = RevealEntry{
				Text:           r.Text,
				Selectors:      r.Selectors,
				Creators:       r.Creators,
				RealAnswer:     r.RealAnswer,
				HouseLie:       r.HouseLie,
				SelectorPoints: r.SelectorPoints,
				CreatorPoints:  r.CreatorPoints,
			}
		}
		room.Broadcast(NewOutgoing(MsgReveal, RevealPayload{Reveals: reveals}))
	case game.ScoreBoard:
		room.Broadcast(NewOutgoing(MsgScores, h.buildPlayerList(room)))
	case game.ScoreBoardFinal:
		room.Broadcast(NewOutgoing(MsgFinalScores, h.buildPlayerList(room)))
	}
}

func (h *Hub) sendCurrentStateData(client *Client, room *Room) {
	snap := room.Game.GetStateSnapshot()
	switch game.State(snap.State) {
	case game.ShowQuestion:
		q := room.Game.GetCurrentQuestion()
		if q != nil {
			client.Send(NewOutgoing(MsgQuestion, QuestionPayload{
				Text:           q.Text,
				QuestionNumber: snap.QuestionIndex + 1,
				TotalQuestions: snap.TotalQuestions,
			}))
		}
	case game.ShowAnswers:
		answers := room.Game.GetAnswersForPlayer(client.UUID)
		options := make([]AnswerOption, len(answers))
		for i, a := range answers {
			options[i] = AnswerOption{Text: a}
		}
		client.Send(NewOutgoing(MsgAnswers, AnswersPayload{Answers: options}))
	case game.RevealTheTruth:
		revealAnswers := room.Game.GetRevealAnswers()
		reveals := make([]RevealEntry, len(revealAnswers))
		for i, r := range revealAnswers {
			reveals[i] = RevealEntry{
				Text:           r.Text,
				Selectors:      r.Selectors,
				Creators:       r.Creators,
				RealAnswer:     r.RealAnswer,
				HouseLie:       r.HouseLie,
				SelectorPoints: r.SelectorPoints,
				CreatorPoints:  r.CreatorPoints,
			}
		}
		client.Send(NewOutgoing(MsgReveal, RevealPayload{Reveals: reveals}))
	case game.ScoreBoard:
		client.Send(NewOutgoing(MsgScores, h.buildPlayerList(room)))
	case game.ScoreBoardFinal:
		client.Send(NewOutgoing(MsgFinalScores, h.buildPlayerList(room)))
	}
}

func (h *Hub) buildGameStateMsg(g *game.Game) []byte {
	snap := g.GetStateSnapshot()
	return NewOutgoing(MsgGameState, GameStatePayload{
		State:          int(snap.State),
		StateTimestamp: snap.StateTimestamp.UnixMilli(),
		StateVersion:   snap.StateVersion,
		RoundIndex:     snap.RoundIndex,
		QuestionIndex:  snap.QuestionIndex,
		TotalQuestions: snap.TotalQuestions,
		Duration:       snap.State.Duration(),
	})
}

func (h *Hub) buildPlayerList(room *Room) PlayerListPayload {
	players := room.Game.GetPlayers()
	list := make([]PlayerPayload, len(players))
	for i, p := range players {
		list[i] = PlayerPayload{
			UUID:     p.UUID,
			Nickname: p.Nickname,
			Score:    p.Score,
			Index:    p.Index,
		}
	}
	return PlayerListPayload{Players: list}
}

// Register adds a client to the hub's register channel.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// sendPersonalizedAnswers sends each client their personalized answer list.
func (h *Hub) sendPersonalizedAnswers(room *Room) {
	room.ForEachClient(func(uuid string, c *Client) {
		answers := room.Game.GetAnswersForPlayer(uuid)
		options := make([]AnswerOption, len(answers))
		for i, a := range answers {
			options[i] = AnswerOption{Text: a}
		}
		c.Send(NewOutgoing(MsgAnswers, AnswersPayload{Answers: options}))
	})
}

// TimeSyncMessage returns a time sync message.
func TimeSyncMessage() []byte {
	return NewOutgoing(MsgTimeSync, TimeSyncPayload{ServerTime: time.Now().UnixMilli()})
}
