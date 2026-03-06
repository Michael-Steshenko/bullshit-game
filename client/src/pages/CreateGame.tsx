import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createGame } from '../lib/api';
import './CreateGame.css';

export function CreateGame() {
  const navigate = useNavigate();
  const [lang, setLang] = useState('en');
  const [totalQ, setTotalQ] = useState(7);
  const [loading, setLoading] = useState(false);

  const handleCreate = async () => {
    setLoading(true);
    try {
      const { pin } = await createGame(lang, totalQ);
      navigate(`/join?pin=${pin}`);
    } catch {
      alert('Failed to create game');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page page-center">
      <div className="card">
        <h2 className="text-center mb-3">Create Game</h2>

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
                className={`btn-secondary q-btn ${totalQ === n ? 'active' : ''}`}
                onClick={() => setTotalQ(n)}
              >
                {n}
              </button>
            ))}
          </div>
        </div>

        <button className="btn-primary full-width" onClick={handleCreate} disabled={loading}>
          {loading ? 'Creating...' : 'Create'}
        </button>
      </div>
    </div>
  );
}
