import { createContext, type Dispatch } from 'react';
import {
  type GameStateData,
  type PlayerData,
  type QuestionData,
  type AnswersData,
  type RevealData,
  type ErrorData,
  type RevealEntry,
  GameState,
} from '../lib/types';

export interface GameStore {
  // Connection
  connected: boolean;

  // Player identity
  myUUID: string;
  myNickname: string;
  myScore: number;
  myIndex: number;

  // Game state
  pin: string;
  validatedPin: string;
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

export type GameAction =
  | { type: 'SET_PIN'; pin: string }
  | { type: 'CREATED_GAME'; data: { pin: string } }
  | { type: 'PIN_VALIDATED'; data: { pin: string } }
  | { type: 'GAME_STATE'; data: GameStateData }
  | {
      type: 'REJOINED';
      data: { uuid: string; nickname: string; score: number; index: number };
    }
  | {
      type: 'PLAYER_JOINED';
      data: { uuid: string; nickname: string; index: number };
    }
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

export const initialState: GameStore = {
  connected: false,
  myUUID: '',
  myNickname: '',
  myScore: 0,
  myIndex: -1,
  pin: '',
  validatedPin: '',
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

export function gameReducer(state: GameStore, action: GameAction): GameStore {
  switch (action.type) {
    case 'SET_PIN':
      return { ...state, pin: action.pin };

    case 'CREATED_GAME':
      return { ...state, pin: action.data.pin, validatedPin: action.data.pin, error: null };

    case 'PIN_VALIDATED':
      return { ...state, pin: action.data.pin, validatedPin: action.data.pin, error: null };

    case 'CONNECTED':
      return { ...state, connected: action.connected };

    case 'GAME_STATE': {
      const d = action.data;
      const newGameState = d.state as GameState;
      const cleared =
        newGameState !== state.gameState ? { submittedPlayers: [], selectedPlayers: [] } : {};
      return {
        ...state,
        ...cleared,
        gameState: newGameState,
        stateVersion: d.stateVersion,
        stateTimestamp: Date.now(),
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
      if (state.players.find((p) => p.uuid === action.data.uuid)) return state;
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
      return {
        ...state,
        submittedPlayers: [...state.submittedPlayers, action.data.uuid],
      };

    case 'ANSWERS':
      return { ...state, answers: action.data.answers };

    case 'ANSWER_SELECTED':
      if (state.selectedPlayers.includes(action.data.uuid)) return state;
      return {
        ...state,
        selectedPlayers: [...state.selectedPlayers, action.data.uuid],
      };

    case 'REVEAL':
      return { ...state, reveals: action.data.reveals };

    case 'SCORES':
    case 'FINAL_SCORES': {
      const me = action.data.players.find((p) => p.uuid === state.myUUID);
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
      return {
        ...state,
        validatedPin: action.data.code === 'GAME_NOT_EXIST' ? '' : state.validatedPin,
        error: action.data,
      };

    case 'RESET':
      return initialState;

    default:
      return state;
  }
}

export interface GameContextType {
  state: GameStore;
  dispatch: Dispatch<GameAction>;
  send: (type: string, payload?: Record<string, unknown>) => void;
}

export const GameContext = createContext<GameContextType | null>(null);
