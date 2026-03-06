import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useGame } from '../context/GameContext';
import { GameState } from '../lib/types';
import './JoinGame.css';

export function JoinGame() {
  const { state, send, dispatch } = useGame();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [pin, setPin] = useState(searchParams.get('pin') || '');
  const [nickname, setNickname] = useState('');
  const [step, setStep] = useState<'pin' | 'nickname'>(searchParams.get('pin') ? 'nickname' : 'pin');

  // Navigate to game when joined
  useEffect(() => {
    if (state.myUUID && state.gameState === GameState.GameStaging) {
      navigate('/game');
    }
  }, [state.myUUID, state.gameState, navigate]);

  const handlePinSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (pin.length === 4) {
      dispatch({ type: 'SET_PIN', pin: pin.toUpperCase() });
      setStep('nickname');
    }
  };

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim()) return;
    const upperPin = pin.toUpperCase();
    dispatch({ type: 'SET_PIN', pin: upperPin });
    send('join', { pin: upperPin, nickname: nickname.trim() });
  };

  return (
    <div className="page page-center">
      <div className="card">
        <h2 className="text-center mb-3">Join Game</h2>

        {step === 'pin' ? (
          <form onSubmit={handlePinSubmit}>
            <div className="form-group mb-3">
              <label>Game PIN</label>
              <input
                type="text"
                value={pin}
                onChange={e => setPin(e.target.value.toUpperCase().slice(0, 4))}
                placeholder="ABCD"
                maxLength={4}
                className="pin-input"
                autoFocus
              />
            </div>
            <button className="btn-primary full-width" disabled={pin.length !== 4}>
              Next
            </button>
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
            {state.error && (
              <p className="error-text">{state.error.message}</p>
            )}
          </form>
        )}
      </div>
    </div>
  );
}
