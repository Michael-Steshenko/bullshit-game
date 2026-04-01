import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useGame } from '../hooks/useGame';
import { GameState } from '../lib/types';
import './CreateGame.css';

export function CreateGame() {
  const { state, send } = useGame();
  const navigate = useNavigate();
  const [nickname, setNickname] = useState('');
  const [lang, setLang] = useState('en');
  const [totalQ, setTotalQ] = useState(7);
  const [createRequested, setCreateRequested] = useState(false);
  const isCreating = createRequested && !state.error && !state.myUUID;

  useEffect(() => {
    if (state.myUUID && state.gameState === GameState.GameStaging && state.pin) {
      navigate('/game');
    }
  }, [navigate, state.gameState, state.myUUID, state.pin]);

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim() || !state.connected) return;
    setCreateRequested(true);
    send('create_and_join', {
      nickname: nickname.trim(),
      lang,
      totalQuestions: totalQ,
    });
  };

  return (
    <div className="page page-center">
      <div className="card">
        <h2 className="text-center mb-3">Create Game</h2>

        <form onSubmit={handleCreate}>
          <div className="form-group mb-2">
            <label>Nickname</label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value.slice(0, 9))}
              placeholder="Your name"
              maxLength={9}
              autoFocus
            />
          </div>

          <div className="form-group mb-2">
            <label>Language</label>
            <select value={lang} onChange={e => setLang(e.target.value)}>
              <option value="en">English</option>
              <option value="he">Hebrew</option>
            </select>
          </div>

          <div className="form-group mb-3">
            <label>Questions</label>
            <div className="question-options">
              {[5, 7, 10].map(n => (
                <button
                  key={n}
                  type="button"
                  className={`btn-secondary q-btn ${totalQ === n ? 'active' : ''}`}
                  onClick={() => setTotalQ(n)}
                >
                  {n}
                </button>
              ))}
            </div>
          </div>

          <button className="btn-primary full-width" disabled={!nickname.trim() || !state.connected || isCreating}>
            {!state.connected ? 'Connecting...' : isCreating ? 'Creating...' : 'Create'}
          </button>
          {state.error && <p className="create-error">{state.error.message}</p>}
        </form>
      </div>
    </div>
  );
}
