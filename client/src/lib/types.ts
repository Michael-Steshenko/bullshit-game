// WebSocket message types
export interface OutgoingMessage {
  type: string;
  data?: unknown;
}

export interface GameStateData {
  state: number;
  stateTimestamp: number;
  stateVersion: number;
  roundIndex: number;
  questionIndex: number;
  totalQuestions: number;
  duration: number;
}

export interface PlayerData {
  uuid: string;
  nickname: string;
  score: number;
  index: number;
}

export interface PlayerListData {
  players: PlayerData[];
}

export interface PlayerJoinedData {
  uuid: string;
  nickname: string;
  index: number;
}

export interface RejoinedData {
  uuid: string;
  nickname: string;
  score: number;
  index: number;
}

export interface QuestionData {
  text: string;
  questionNumber: number;
  totalQuestions: number;
}

export interface AnswerSubmittedData {
  uuid: string;
}

export interface AnswersData {
  answers: { text: string }[];
}

export interface AnswerSelectedData {
  uuid: string;
}

export interface RevealEntry {
  text: string;
  selectors: string[];
  creators: string[];
  realAnswer: boolean;
  houseLie: boolean;
  selectorPoints: number;
  creatorPoints: number;
}

export interface RevealData {
  reveals: RevealEntry[];
}

export interface ScoresData {
  players: PlayerData[];
}

export interface ErrorData {
  code: string;
  message: string;
}

export interface TimeSyncData {
  serverTime: number;
}

// Game states
export const GameState = {
  GameStaging: 0,
  RoundIntro: 1,
  ShowQuestion: 2,
  ShowAnswers: 3,
  RevealTheTruth: 4,
  ScoreBoard: 5,
  ScoreBoardFinal: 6,
} as const;

export type GameState = (typeof GameState)[keyof typeof GameState];

export function gameStateName(state: GameState): string {
  switch (state) {
    case GameState.GameStaging: return 'Lobby';
    case GameState.RoundIntro: return 'Round Intro';
    case GameState.ShowQuestion: return 'Question';
    case GameState.ShowAnswers: return 'Voting';
    case GameState.RevealTheTruth: return 'Reveal';
    case GameState.ScoreBoard: return 'Scores';
    case GameState.ScoreBoardFinal: return 'Final Scores';
    default: return 'Unknown';
  }
}
