import { useNavigate } from 'react-router-dom';
import './Landing.css';

export function Landing() {
  const navigate = useNavigate();

  return (
    <div className="page page-center">
      <div className="landing">
        <h1 className="landing-title">💩 Bullshit</h1>
        <p className="landing-subtitle">The party game of creative lying</p>

        <div className="landing-buttons">
          <button className="btn-primary landing-btn" onClick={() => navigate('/create')}>
            Create Game
          </button>
          <button className="btn-secondary landing-btn" onClick={() => navigate('/join')}>
            Join Game
          </button>
        </div>
      </div>
    </div>
  );
}
