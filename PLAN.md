# Bullshit.wtf Clone -- Rewrite Plan

## Tech Stack

- **Frontend**: React (Vite + TypeScript)
- **Backend**: Go
- **Database**: Postgres (questions, game history only -- NOT live game state)
- **Real-time**: gorilla/websocket (server) + native WebSocket API (browser)
- **Auth**: Ephemeral in-memory UUID + nickname (no DB persistence). OAuth (Apple/Google) added later.
- **Containerization**: Docker Compose on OrbStack
- **Reverse proxy / TLS**: Caddy with automatic Let's Encrypt
- **Domain**: bullshit.steshenko.net

---

## Game Overview

Fibbage/Balderdash-style multiplayer party game. Players are shown trivia questions with
missing information. Everyone writes a convincing fake answer. All answers (player lies,
system-generated "house lies", and the real truth) are shuffled and displayed. Players
pick the one they think is real. Points for finding the truth, points for fooling others,
penalties for falling for house lies.

---

## Game States

| State               | ID  | Duration                             | Trigger to advance                                   |
| ------------------- | --- | ------------------------------------ | ---------------------------------------------------- |
| GameStaging (lobby) | 0   | No timer                             | Manual (room creator only)                           |
| RoundIntro          | 1   | 5 seconds                            | Auto-timer                                           |
| ShowQuestion        | 2   | 25 seconds                           | Auto-timer OR all players submitted                  |
| ShowAnswers         | 3   | 20 seconds                           | Auto-timer OR all players selected                   |
| RevealTheTruth      | 4   | No timer (7s per answer, auto-steps) | Auto after last reveal                               |
| ScoreBoard          | 5   | 5 seconds                            | Auto-timer                                           |
| ScoreBoardFinal     | 6   | Terminal                             | Host rematch resets current game in-place (same PIN) |

### State Transition Flow

```
GameStaging -> RoundIntro -> ShowQuestion -> ShowAnswers -> RevealTheTruth -> ScoreBoard
                    ^                                                             |
                    |_____________ (next question, same round) __________________|
                    |_____________ (next round, roundIndex++) ___________________|
                                                                                  |
                                                              ScoreBoardFinal <--- (last question)
```

### Round Structure

The game divides `totalQ` questions into 3 rounds with escalating points:

- Round 0: questions 0 through floor(totalQ/2) - 1
- Round 1: questions floor(totalQ/2) through totalQ - 2
- Round 2: the final question (totalQ - 1)

Example with 7 questions: Round 0 = questions 0-2, Round 1 = questions 3-5, Round 2 = question 6.

---

## Scoring Rules

Points scale by round index (0, 1, 2):

| Scenario                                         | Round 0                          | Round 1 | Round 2 |
| ------------------------------------------------ | -------------------------------- | ------- | ------- |
| Player selects the real answer                   | +1000                            | +1500   | +2000   |
| Player fools someone with their lie (per victim) | +500                             | +750    | +1000   |
| Player selects a house lie                       | -500                             | -750    | -1000   |
| Player selects another player's lie              | 0 (creator gets bullshit points) | 0       | 0       |

Scores are cumulative across all questions.

---

## Question Format

```
id: int
questionText: string      -- contains "$blank$" placeholder for fill-in-the-blank
realAnswer: string         -- the correct answer
fakeAnswers: string[]      -- pre-authored "house lies" (system decoys)
citation: string           -- source reference for the correct answer
lang: string               -- language code ("en", "he", etc.)
```

### House Lie Logic

House lies fill the gap between unique player answers and player count:

- houseLiesNeeded = playerCount - uniquePlayerAnswerCount
- Taken from the question's fakeAnswers array in order
- If a player doesn't submit: their slot is filled by a house lie
- If two players submit identical text: one house lie fills the duplicate slot

The real answer is always added to the answer pool.

---

## Answer Rules

- All player answers are lowercased for storage and comparison
- If a player submits the actual correct answer, the server rejects it (they must try again)
- Players cannot see/select their own answer in the voting phase
- Duplicate answers between players are allowed but deduplicated in display
- Server normalizes answer text as `strings.ToLower(strings.TrimSpace(text))`
- Empty or whitespace-only answers are rejected with `EMPTY_ANSWER`
- Answers longer than 40 characters are rejected with `ANSWER_TOO_LONG`
- A player may re-submit during ShowQuestion; latest submission for that player overwrites their previous text
- If multiple players submit the same normalized answer, all are accepted; ShowAnswers shows one option for that text, and house-lie fill still follows `houseLiesNeeded = playerCount - uniquePlayerAnswerCount`

---

## Player Rules

- Max 8 players per game
- Nicknames: max 9 characters, no duplicate check
- Game PIN: 4 uppercase letters
- No minimum player count enforced

---

## Roles

### Player

- Picks a nickname to join
- Submits fake answers
- Votes on answers
- Sees their own score in the footer
- The player who created the room is the "host" and is the only one who can start the game

### Presenter (Phase 2 -- not in initial build)

- Dedicated display device (TV, projector, second screen)
- Does NOT play -- no nickname, no answers, no votes
- Controls the game clock (only presenter triggers state transitions when present)
- Plays all sound effects
- Shows large "Join at {domain} with PIN: XXXX" banner in lobby
- No game footer (no player name/score bar)

---

## Auth Strategy

### Phase 1 (now)

- Player picks a nickname, server generates an in-memory UUID tied to their game player slot
- Browser stores the UUID + nickname in sessionStorage
- On reconnect (refresh/socket drop), client sends UUID + PIN and the server rebinds to the same player slot if the game is still active
- UUID is discarded only when the game ends or the player explicitly leaves
- No database records for anonymous players

### Phase 2 (later)

- Add Sign in with Apple / Google OAuth
- First OAuth login creates a user row in Postgres
- If a player is mid-game with a temp UUID when they authenticate, link the UUID to their new user record
- Authenticated users get persistent nicknames, game history, stats

---

## Architecture

### Go Backend

```
server/
  cmd/
    server/
      main.go              -- entry point, HTTP server setup
  internal/
    config/
      config.go            -- env vars, DB connection string, etc.
    db/
      db.go                -- Postgres connection pool
      migrations/          -- SQL migration files
      queries/             -- SQL query files (or sqlc generated)
    game/
      state.go             -- GameState enum, state machine logic
      game.go              -- Game struct, in-memory game management
      scoring.go           -- score calculation
      questions.go         -- question fetching and randomization
      pin.go               -- PIN generation
    hub/
      hub.go               -- WebSocket hub (manages all connections)
      room.go              -- per-game room (broadcasts to players in a game)
      client.go            -- single WebSocket connection handler
      messages.go          -- message type definitions (incoming/outgoing)
    handlers/
      http.go              -- HTTP routes (create game, health check, etc.)
      ws.go                -- WebSocket upgrade handler
  go.mod
  go.sum
```

### In-Memory Game State

Live game state lives in Go server memory, NOT in Postgres:

```go
type Game struct {
    PIN             string
    State           GameState
    StateTimestamp  time.Time
    RoundIndex      int
    QuestionIndex   int
    TotalQuestions  int
    Lang            string
    CurrentQuestion *Question
    Players         map[string]*Player    // keyed by temp UUID
    HostID          string                // UUID of the player who created the room
    Answers         map[string]*Answer    // keyed by player UUID or generated ID
    Selections      map[string]*Selection // keyed by player UUID
    RevealAnswers   []RevealAnswer        // computed when entering RevealTheTruth state.
                                          // one entry per unique answer (player lies, house lies, real answer).
                                          // each entry contains: answer text, creator IDs, selector IDs,
                                          // realAnswer/houseLie flags, and points earned/lost.
                                          // sorted alphabetically with the real answer always last.
                                          // sent to clients so the frontend can step through the
                                          // animated reveal sequence without further server calls.
    HasPresenter    bool                  // Phase 2: presenter support
    QuestionIDs     []int                 // pre-selected question IDs for this game
}
```

### Timer Strategy: Client-Side

Timers are owned by the client, not the server. This matches the original design and is
fairer: every player gets the same local countdown regardless of network latency. A player
with a slow connection won't lose time waiting for server round-trips.

How it works:

- Server broadcasts a state change with a `duration` field (e.g. 25000ms for ShowQuestion)
- Each client starts a local countdown
- When a client's timer expires, it sends a `tick` message to the server
- The server processes the FIRST tick it receives and ignores duplicates

The server does NOT set its own timers for gameplay phases. If all clients disconnect,
the game simply pauses in its current state until someone reconnects.

### Transition Idempotency

To keep transitions deterministic and avoid duplicate processing:

- Server stores `stateVersion` on each game and increments it on every successful transition
- `submit_answer`, `select_answer`, and `tick` include the sender's current `stateVersion`
- Messages with stale `stateVersion` are ignored as out-of-date/duplicate

### React Frontend

```
client/
  src/
    main.tsx
    App.tsx
    routes.tsx                  -- React Router setup
    hooks/
      useWebSocket.ts           -- WebSocket connection hook
      useGameState.ts           -- game state from WebSocket
      useSession.ts             -- player session (sessionStorage)
      useTimer.ts               -- countdown timer display
    pages/
      Landing.tsx               -- mobile/desktop detection, menu
      CreateGame.tsx            -- language + length picker
      JoinGame.tsx              -- PIN entry -> nickname entry
      Learn.tsx                 -- how to play
      GameStaging.tsx           -- lobby, player list, START button
      RoundIntro.tsx            -- round splash (5s)
      ShowQuestion.tsx          -- question display + answer input
      ShowAnswers.tsx           -- answer voting
      RevealTheTruth.tsx        -- animated reveal sequence
      ScoreBoard.tsx            -- mid-game scores
      ScoreBoardFinal.tsx       -- final scores + REPLAY
    components/
      GameHeader.tsx            -- top bar (home, PIN, action button)
      GameFooter.tsx            -- bottom bar (nickname, score)
      ProgressBar.tsx           -- countdown bar with panic mode
      PlayerCard.tsx            -- avatar + nickname display
      LeaveGameModal.tsx        -- "Don't leave us" confirmation
    context/
      GameContext.tsx            -- game state provider
      SessionContext.tsx         -- player session provider
    lib/
      sounds.ts                 -- use-sound hook wrappers for game audio
      api.ts                    -- HTTP API calls (create game, etc.)
    assets/
      avatars/                  -- 8 avatar images
      sounds/                   -- all sound files
    styles/
      ...
  index.html
  vite.config.ts
  tsconfig.json
  package.json
```

---

## WebSocket Protocol

All messages are JSON. Direction: C = client->server, S = server->client.

### Client -> Server Messages

| Type            | Payload                       | When                                                                  |
| --------------- | ----------------------------- | --------------------------------------------------------------------- |
| `join`          | `{ pin, nickname }`           | Player joins a game                                                   |
| `reconnect`     | `{ pin, uuid, nickname }`     | Client reconnects to an active game using session identity            |
| `start_game`    | `{ pin }`                     | Host (room creator) starts the game                                   |
| `submit_answer` | `{ pin, text, stateVersion }` | Player submits a fake answer                                          |
| `select_answer` | `{ pin, text, stateVersion }` | Player votes for an answer                                            |
| `tick`          | `{ pin, stateVersion }`       | Client requests state advance (RevealTheTruth -> ScoreBoard)          |
| `rematch`       | `{ pin }`                     | Host requests a rematch (reset game, new questions, same players/PIN) |

### Server -> Client Messages

| Type               | Payload                                                                              | When                                                                           |
| ------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| `game_state`       | `{ state, stateTimestamp, stateVersion, roundIndex, questionIndex, totalQuestions }` | Every state change                                                             |
| `rejoined`         | `{ uuid, nickname, score, index }`                                                   | Reconnect succeeded; existing player slot rebound                              |
| `player_joined`    | `{ uuid, nickname, index }`                                                          | A player joins the lobby                                                       |
| `player_list`      | `{ players: [{ uuid, nickname, score, index }] }`                                    | Full player list (on join)                                                     |
| `question`         | `{ text, questionNumber, totalQuestions }`                                           | ShowQuestion begins                                                            |
| `answer_submitted` | `{ uuid }`                                                                           | A player submitted (for progress tracking)                                     |
| `answers`          | `{ answers: [{ text }] }`                                                            | ShowAnswers begins (shuffled, player's own excluded)                           |
| `answer_selected`  | `{ uuid }`                                                                           | A player voted (for progress tracking)                                         |
| `reveal`           | `{ reveals: [{ text, selectors, creators, realAnswer, houseLie, points }] }`         | RevealTheTruth begins                                                          |
| `scores`           | `{ players: [{ uuid, nickname, score, index }] }`                                    | ScoreBoard data                                                                |
| `final_scores`     | `{ players: [{ uuid, nickname, score, index }] }`                                    | ScoreBoardFinal data                                                           |
| `rematch`          | `{}`                                                                                 | Game reset: new questions, scores zeroed, back to lobby. Same PIN and players. |
| `error`            | `{ code, message }`                                                                  | Error (game full, correct answer, etc.)                                        |
| `time_sync`        | `{ serverTime }`                                                                     | Clock synchronization                                                          |

### Reconnect Contract

On WebSocket connect:

1. If browser has `{ pin, uuid, nickname }` in sessionStorage, send `reconnect`.
2. Server checks game + player slot existence and rebinds this socket to that UUID.
3. Server sends, in order: `rejoined`, `player_list`, `game_state`, and current state payload (`question`, `answers`, `reveal`, `scores`, or `final_scores`).
4. If reconnect fails (`RECONNECT_FAILED`), client falls back to `join`.

### Rematch Semantics (Same PIN)

- Only the host can trigger rematch, and only from `ScoreBoardFinal`.
- Rematch keeps the same PIN and reuses the current in-memory `Game` struct (no new room/PIN).
- Existing players and UUIDs are preserved; all scores are reset to 0.
- Server resets mutable round fields in-place: `state`, `stateTimestamp`, `stateVersion=0`, `roundIndex`, `questionIndex`, answers, selections, and reveal data.
- Server re-rolls question IDs for the new match and broadcasts `rematch` to all connected clients.

### Error Codes (Core)

- `EMPTY_ANSWER`
- `ANSWER_TOO_LONG`
  `RECONNECT_FAILED`
- `CORRECT_ANSWER`
- `GAME_IS_FULL`
- `GAME_NOT_EXIST`

---

## Database Schema (Postgres)

Only persistent data goes in Postgres. Live game state is in-memory.

### questions

```sql
CREATE TABLE questions (
    id          SERIAL PRIMARY KEY,
    lang        TEXT NOT NULL DEFAULT 'en',
    question    TEXT NOT NULL,
    real_answer TEXT NOT NULL,
    fake_answers TEXT[] NOT NULL,     -- array of house lies
    citation    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_lang ON questions(lang);
```

### game_history (Phase 2 -- only after OAuth is implemented)

Only games where at least one player is authenticated (signed in via Apple/Google)
are written to this table. Fully anonymous games are not persisted.

```sql
CREATE TABLE game_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pin         TEXT NOT NULL,
    lang        TEXT NOT NULL,
    total_q     INT NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL,
    ended_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    players     JSONB NOT NULL        -- [{ nickname, score, rank, user_id? }]
);
```

### users (Phase 2 -- OAuth only, not created yet)

```sql
-- Phase 2: only created when a user authenticates via Apple/Google
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider        TEXT NOT NULL,           -- 'apple', 'google'
    provider_id     TEXT NOT NULL,           -- OAuth subject ID
    email           TEXT,
    display_name    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(provider, provider_id)
);
```

---

## PIN Generation

4-character uppercase alphabetic PIN, generated from an atomic counter:

1. Increment a global counter (in-memory atomic int64, seeded from max game_history count on startup)
2. Convert to base-26 using letters A-Z
3. Left-pad with 'A' to exactly 4 characters

---

## Sound Effects

All sounds played on every client using use-sound (React hook wrapper around Howler.js):

| Sound                | When                      | Behavior                          |
| -------------------- | ------------------------- | --------------------------------- |
| staging.mp3          | Lobby                     | Loop, fade out 1s on leave        |
| during-game.mp3      | ShowQuestion, ShowAnswers | Play once, replaced at 5s warning |
| time-warning.mp3     | <5 seconds remaining      | Replaces during-game.mp3          |
| the-truth.mp3        | Reveal: real answer shown | Play once                         |
| house-lie-{0,1}.mp3  | Reveal: house lie shown   | Random variant, play once         |
| player-lie-{0,1}.mp3 | Reveal: player lie shown  | Random variant, play once         |
| final.mp3            | ScoreBoardFinal           | Loop, fade out 1s on leave        |

---

## Reveal Animation Sequence

For each answer in the reveal array (sorted alphabetically, real answer last):

1. t=0s: Show the answer text centered. Show voter avatars below.
2. t=3s: Reveal who wrote it above the answer. Play sound effect. Show points.
3. t=7s: Advance to next answer.

After the last answer, auto-advance to ScoreBoard.

---

## Docker Compose Setup

```yaml
services:
  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
    depends_on:
      - server

  server:
    build: ./server
    environment:
      - DATABASE_URL=postgres://bullshit:bullshit@db:5432/bullshit?sslmode=disable
      - PORT=8080
    depends_on:
      - db

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=bullshit
      - POSTGRES_PASSWORD=bullshit
      - POSTGRES_DB=bullshit
    volumes:
      - pg_data:/var/lib/postgresql/data

volumes:
  caddy_data:
  pg_data:
```

The React frontend is built as static files and served by the Go server (or by Caddy
directly from a volume). No separate container needed for the frontend.

### Caddyfile

```
bullshit.steshenko.net {
    reverse_proxy server:8080
}
```

Caddy handles automatic HTTPS via Let's Encrypt. No manual cert management.

---

## Project Structure

```
bullshit-wtf-clone/
  PLAN.md               -- this file
  docker-compose.yml
  Caddyfile
  server/
    cmd/server/main.go
    internal/
      config/
      db/
        migrations/
        queries/
      game/
      hub/
      handlers/
    go.mod
    Dockerfile
  client/
    src/
      ...
    package.json
    vite.config.ts
    Dockerfile          -- multi-stage: build with node, serve static files
```

---

## Implementation Phases

### Phase 1: Core Game Loop

1. Go server with WebSocket hub and in-memory game state machine
2. Protocol idempotency with `stateVersion` on `submit_answer`, `select_answer`, and `tick`
3. Reconnection/session restore using UUID + PIN rebind to existing player slot
4. Postgres with questions table and seed data
5. React frontend with all game pages
6. Tests: Jest for client/protocol behavior + Go `testing` package for game engine/scoring
7. Docker Compose with Caddy (after core gameplay loop and tests are passing locally)

### Phase 2: Polish & Presenter

1. Sound effects
2. Reveal animations
3. Mobile/desktop responsive landing page
4. Progress bars with panic mode
5. Leave game confirmation modal
6. Presenter mode (dedicated display device for shared screen)

### Phase 3: Persistence & Auth

1. OAuth (Apple/Google) with user table
2. Game history written to Postgres on game completion (only if at least one player is authenticated)
3. Player profiles and stats

### Phase 4: Enhancements (optional)

1. Custom question packs
2. Spectator mode
3. Game replays
4. Leaderboards
