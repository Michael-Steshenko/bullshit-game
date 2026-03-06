import React, { createContext, useContext, useReducer, useCallback, useEffect, type ReactNode } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { useSession } from '../hooks/useSession';
import type {
  GameStateData, PlayerData, QuestionData, AnswersData,
  RevealData, ErrorData, RevealEntry,
} from '../lib/types';
import { GameState } from '../lib/types';

interface GameStore {
  // Connection
  connected: boolean;

  // Player identity
  myUUID: string;
  myNickname: string;
  myScore: number;
  myIndex: number;

  // Game state
  pin: string;
  gameState: GameState;
  stateVersion: number;
  stateTimestamp: number;
  roundIndex: number;
  questionIndex: number;
  totalQuestions: number;
  duration: number;

  // Players
  players: PlayerData[];

  // Round data
  question: QuestionData | null;
  answers: { text: string }[];
  reveals: RevealEntry[];
  submittedPlayers: string[];
  selectedPlayers: string[];

  // Error
  error: ErrorData | null;

  // Host
  isHost: boolean;
}

type GameAction =
  | { type: 'SET_PIN'; pin: string }
  | { type: 'GAME_STATE'; data: GameStateData }
  | { type: 'REJOINED'; data: { uuid: string; nickname: string; score: number; index: number } }
  | { type: 'PLAYER_JOINED'; data: { uuid: string; nickname: string; index: number } }
  | { type: 'PLAYER_LIST'; data: { players: PlayerData[] } }
  | { type: 'QUESTION'; data: QuestionData }
  | { type: 'ANSWER_SUBMITTED'; data: { uuid: string } }
  | { type: 'ANSWERS'; data: AnswersData }
  | { type: 'ANSWER_SELECTED'; data: { uuid: string } }
  | { type: 'REVEAL'; data: RevealData }
  | { type: 'SCORES'; data: { players: PlayerData[] } }
  | { type: 'FINAL_SCORES'; data: { players: PlayerData[] } }
  | { type: 'REMATCH' }
  | { type: 'ERROR'; data: ErrorData }
  | { type: 'CONNECTED'; connected: boolean }
  | { type: 'RESET' };

const initialState: GameStore = {
  connected: false,
  myUUID: '',
  myNickname: '',
  myScore: 0,
  myIndex: -1,
  pin: '',
  gameState: GameState.GameStaging,
  stateVersion: 0,
  stateTimestamp: 0,
  roundIndex: 0,
  questionIndex: 0,
  totalQuestions: 0,
  duration: 0,
  players: [],
  question: null,
  answers: [],
  reveals: [],
  submittedPlayers: [],
  selectedPlayers: [],
  error: null,
  isHost: false,
};

function gameReducer(state: GameStore, action: GameAction): GameStore {
  switch (action.type) {
    case 'SET_PIN':
      return { ...state, pin: action.pin };

    case 'CONNECTED':
      return { ...state, connected: action.connected };

    case 'GAME_STATE': {
      const d = action.data;
      const newGameState = d.state as GameState;
      const cleared = newGameState !== state.gameState
        ? { submittedPlayers: [], selectedPlayers: [] }
        : {};
      return {
        ...state,
        ...cleared,
        gameState: newGameState,
        stateVersion: d.stateVersion,
        stateTimestamp: d.stateTimestamp,
        roundIndex: d.roundIndex,
        questionIndex: d.questionIndex,
        totalQuestions: d.totalQuestions,
        duration: d.duration,
      };
    }

    case 'REJOINED':
      return {
        ...state,
        myUUID: action.data.uuid,
        myNickname: action.data.nickname,
        myScore: action.data.score,
        myIndex: action.data.index,
        isHost: action.data.index === 0,
      };

    case 'PLAYER_JOINED':
      if (state.players.find(p => p.uuid === action.data.uuid)) return state;
      return {
        ...state,
        players: [...state.players, { ...action.data, score: 0 }],
      };

    case 'PLAYER_LIST':
      return {
        ...state,
        players: action.data.players,
        isHost: action.data.players.length > 0 && action.data.players[0].uuid === state.myUUID,
      };

    case 'QUESTION':
      return { ...state, question: action.data };

    case 'ANSWER_SUBMITTED':
      if (state.submittedPlayers.includes(action.data.uuid)) return state;
      return { ...state, submittedPlayers: [...state.submittedPlayers, action.data.uuid] };

    case 'ANSWERS':
      return { ...state, answers: action.data.answers };

    case 'ANSWER_SELECTED':
      if (state.selectedPlayers.includes(action.data.uuid)) return state;
      return { ...state, selectedPlayers: [...state.selectedPlayers, action.data.uuid] };

    case 'REVEAL':
      return { ...state, reveals: action.data.reveals };

    case 'SCORES':
    case 'FINAL_SCORES': {
      const me = action.data.players.find(p => p.uuid === state.myUUID);
      return {
        ...state,
        players: action.data.players,
        myScore: me?.score ?? state.myScore,
      };
    }

    case 'REMATCH':
      return {
        ...state,
        gameState: GameState.GameStaging,
        stateVersion: 0,
        roundIndex: 0,
        questionIndex: 0,
        question: null,
        answers: [],
        reveals: [],
        submittedPlayers: [],
        selectedPlayers: [],
        error: null,
      };

    case 'ERROR':
      return { ...state, error: action.data };

    case 'RESET':
      return initialState;

    default:
      return state;
  }
}

interface GameContextType {
  state: GameStore;
  dispatch: React.Dispatch<GameAction>;
  send: (type: string, payload?: Record<string, unknown>) => void;
}

const GameContext = createContext<GameContextType | null>(null);

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(gameReducer, initialState);
  const { getSession, setSession } = useSession();

  const handleMessage = useCallback((msg: { type: string; data: unknown }) => {
    switch (msg.type) {
      case 'game_state':
        dispatch({ type: 'GAME_STATE', data: msg.data as GameStateData });
        break;
      case 'rejoined':
        dispatch({ type: 'REJOINED', data: msg.data as GameStore['players'][0] & { score: number } });
        break;
      case 'player_joined':
        dispatch({ type: 'PLAYER_JOINED', data: msg.data as { uuid: string; nickname: string; index: number } });
        break;
      case 'player_list':
        dispatch({ type: 'PLAYER_LIST', data: msg.data as { players: PlayerData[] } });
        break;
      case 'question':
        dispatch({ type: 'QUESTION', data: msg.data as QuestionData });
        break;
      case 'answer_submitted':
        dispatch({ type: 'ANSWER_SUBMITTED', data: msg.data as { uuid: string } });
        break;
      case 'answers':
        dispatch({ type: 'ANSWERS', data: msg.data as AnswersData });
        break;
      case 'answer_selected':
        dispatch({ type: 'ANSWER_SELECTED', data: msg.data as { uuid: string } });
        break;
      case 'reveal':
        dispatch({ type: 'REVEAL', data: msg.data as RevealData });
        break;
      case 'scores':
        dispatch({ type: 'SCORES', data: msg.data as { players: PlayerData[] } });
        break;
      case 'final_scores':
        dispatch({ type: 'FINAL_SCORES', data: msg.data as { players: PlayerData[] } });
        break;
      case 'rematch':
        dispatch({ type: 'REMATCH' });
        break;
      case 'error':
        dispatch({ type: 'ERROR', data: msg.data as ErrorData });
        break;
      case 'time_sync':
        // Could store time offset if needed
        break;
    }
  }, []);

  const { connected, send } = useWebSocket(handleMessage);

  // Attempt reconnect on mount
  useEffect(() => {
    if (connected) {
      const session = getSession();
      if (session) {
        send('reconnect', { pin: session.pin, uuid: session.uuid, nickname: session.nickname });
        dispatch({ type: 'SET_PIN', pin: session.pin });
      }
    }
  }, [connected, getSession, send]);

  // Sync connected state
  useEffect(() => {
    dispatch({ type: 'CONNECTED', connected });
  }, [connected]);

  // Save session when identity is set
  useEffect(() => {
    if (state.myUUID && state.pin && state.myNickname) {
      setSession({ pin: state.pin, uuid: state.myUUID, nickname: state.myNickname });
    }
  }, [state.myUUID, state.pin, state.myNickname, setSession]);

  return (
    <GameContext.Provider value={{ state, dispatch, send }}>
      {children}
    </GameContext.Provider>
  );
}

export function useGame() {
  const ctx = useContext(GameContext);
  if (!ctx) throw new Error('useGame must be used within GameProvider');
  return ctx;
}
