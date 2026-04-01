import { useState, useEffect, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useGame } from '../hooks/useGame';
import { GameState } from '../lib/types';
import './JoinGame.css';

export function JoinGame() {
  const { state, send } = useGame();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [pin, setPin] = useState(searchParams.get('pin')?.toUpperCase() || '');
  const [nickname, setNickname] = useState('');
  const validatedPin = state.validatedPin;
  const normalizedPin = pin.toUpperCase();
  const pinValidated = normalizedPin.length === 4 && validatedPin === normalizedPin;
  const pinError = useMemo(() => {
    if (state.error?.code === 'GAME_NOT_EXIST') return state.error.message;
    return '';
  }, [state.error]);

  // Navigate to game when joined
  useEffect(() => {
    if (state.myUUID && state.gameState === GameState.GameStaging) {
      navigate('/game');
    }
  }, [state.myUUID, state.gameState, navigate]);

  const handleValidatePin = (e: React.FormEvent) => {
    e.preventDefault();
    if (pin.length !== 4 || !state.connected) return;
    send('validate_pin', { pin: normalizedPin });
  };

  const handlePinChange = (value: string) => {
    const next = value.toUpperCase().slice(0, 4);
    setPin(next);
  };

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim() || !state.connected) return;
    if (!pinValidated) {
      send('validate_pin', { pin: normalizedPin });
      return;
    }
    send('join', { pin: normalizedPin, nickname: nickname.trim() });
  };

  return (
    <div className="page page-center">
      <div className="card">
        <h2 className="text-center mb-3">Join Game</h2>

        {!pinValidated ? (
          <form onSubmit={handleValidatePin}>
            <div className="form-group mb-3">
              <label>Game PIN</label>
              <input
                type="text"
                value={pin}
                onChange={e => handlePinChange(e.target.value)}
                placeholder="ABCD"
                maxLength={4}
                className="pin-input"
                autoFocus
              />
            </div>
            <button className="btn-primary full-width" disabled={pin.length !== 4 || !state.connected}>
              Next
            </button>
            {pinError && <p className="error-text">{pinError}</p>}
          </form>
        ) : (
          <form onSubmit={handleJoin}>
            <div className="pin-display mb-2">PIN: {pin}</div>
            <div className="form-group mb-3">
              <label>Nickname</label>
              <input
                type="text"
                value={nickname}
                onChange={e => setNickname(e.target.value.slice(0, 9))}
                placeholder="Your name"
                maxLength={9}
                autoFocus
              />
            </div>
            <button className="btn-primary full-width" disabled={!nickname.trim() || !state.connected}>
              {state.connected ? 'Join' : 'Connecting...'}
            </button>
            {state.error && <p className="error-text">{state.error.message}</p>}
            <button
              type="button"
              className="btn-secondary full-width back-btn"
              onClick={() => setPin('')}
            >
              Change PIN
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
