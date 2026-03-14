import { useNavigate } from 'react-router-dom';
import { useGame } from '../hooks/useGame';
import './GameHeader.css';

export function GameHeader() {
  const { state } = useGame();
  const navigate = useNavigate();

  return (
    <header className="game-header">
      <button className="header-home" onClick={() => navigate('/')}>
        🏠
      </button>
      {state.pin && (
        <div className="header-pin">
          PIN: <strong>{state.pin}</strong>
        </div>
      )}
    </header>
  );
}
