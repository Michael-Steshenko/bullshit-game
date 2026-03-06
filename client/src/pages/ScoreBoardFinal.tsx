import { useGame } from '../context/GameContext';
import './ScoreBoardFinal.css';

export function ScoreBoardFinal() {
  const { state, send } = useGame();

  const handleRematch = () => {
    send('rematch', { pin: state.pin });
  };

  const sorted = [...state.players].sort((a, b) => b.score - a.score);
  const winner = sorted[0];

  return (
    <div className="final-scoreboard fade-in">
      <h1 className="mb-2">🏆 Game Over!</h1>

      {winner && (
        <div className="winner-banner mb-3">
          <span className="winner-label">Winner</span>
          <span className="winner-name">{winner.nickname}</span>
          <span className="winner-score">{winner.score.toLocaleString()} pts</span>
        </div>
      )}

      <div className="score-list mb-3">
        {sorted.map((p, i) => (
          <div key={p.uuid} className={`score-row ${p.uuid === state.myUUID ? 'me' : ''}`}>
            <span className="score-rank">#{i + 1}</span>
            <span className="score-name">{p.nickname}</span>
            <span className="score-value">{p.score.toLocaleString()}</span>
          </div>
        ))}
      </div>

      {state.isHost && (
        <button className="btn-primary" onClick={handleRematch}>
          Play Again
        </button>
      )}
    </div>
  );
}
