import { useReducer, useCallback, useEffect, type ReactNode } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { useSession } from '../hooks/useSession';
import type {
  CreatedGameData,
  GameStateData,
  PlayerData,
  PinValidatedData,
  QuestionData,
  AnswersData,
  RevealData,
  ErrorData,
} from '../lib/types';
import { type GameStore, initialState, gameReducer, GameContext } from './GameContext.types';

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(gameReducer, initialState);
  const { getSession, setSession } = useSession();

  const handleMessage = useCallback((msg: { type: string; data: unknown }) => {
    switch (msg.type) {
      case 'game_state':
        dispatch({ type: 'GAME_STATE', data: msg.data as GameStateData });
        break;
      case 'created_game':
        dispatch({ type: 'CREATED_GAME', data: msg.data as CreatedGameData });
        break;
      case 'pin_validated':
        dispatch({ type: 'PIN_VALIDATED', data: msg.data as PinValidatedData });
        break;
      case 'rejoined':
        dispatch({
          type: 'REJOINED',
          data: msg.data as GameStore['players'][0] & { score: number },
        });
        break;
      case 'player_joined':
        dispatch({
          type: 'PLAYER_JOINED',
          data: msg.data as { uuid: string; nickname: string; index: number },
        });
        break;
      case 'player_list':
        dispatch({
          type: 'PLAYER_LIST',
          data: msg.data as { players: PlayerData[] },
        });
        break;
      case 'question':
        dispatch({ type: 'QUESTION', data: msg.data as QuestionData });
        break;
      case 'answer_submitted':
        dispatch({
          type: 'ANSWER_SUBMITTED',
          data: msg.data as { uuid: string },
        });
        break;
      case 'answers':
        dispatch({ type: 'ANSWERS', data: msg.data as AnswersData });
        break;
      case 'answer_selected':
        dispatch({
          type: 'ANSWER_SELECTED',
          data: msg.data as { uuid: string },
        });
        break;
      case 'reveal':
        dispatch({ type: 'REVEAL', data: msg.data as RevealData });
        break;
      case 'scores':
        dispatch({
          type: 'SCORES',
          data: msg.data as { players: PlayerData[] },
        });
        break;
      case 'final_scores':
        dispatch({
          type: 'FINAL_SCORES',
          data: msg.data as { players: PlayerData[] },
        });
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

  const { connected, send: wsSend } = useWebSocket(handleMessage);

  const send = useCallback(
    (type: string, payload?: Record<string, unknown>) => {
      dispatch({ type: 'CLEAR_ERROR' });
      wsSend(type, payload);
    },
    [wsSend]
  );

  // Attempt reconnect on mount
  useEffect(() => {
    if (connected) {
      const session = getSession();
      if (session) {
        send('reconnect', {
          pin: session.pin,
          uuid: session.uuid,
          nickname: session.nickname,
        });
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
      setSession({
        pin: state.pin,
        uuid: state.myUUID,
        nickname: state.myNickname,
      });
    }
  }, [state.myUUID, state.pin, state.myNickname, setSession]);

  return <GameContext.Provider value={{ state, dispatch, send }}>{children}</GameContext.Provider>;
}
