import { useGame } from '../hooks/useGame';
import { getAvatarForPlayerIndex } from '../lib/avatar';
import './ScoreBoardFinal.css';

export function ScoreBoardFinal() {
  const { state, send } = useGame();

  const handleRematch = () => {
    send('rematch', { pin: state.pin });
  };

  const sorted = [...state.players].sort((a, b) => b.score - a.score);
  const maxScore = sorted.length > 0 ? sorted[0].score : -1;
  const winners = sorted.filter((p) => p.score === maxScore && maxScore >= 0);

  return (
    <div className="final-scoreboard fade-in">
      <h1 className="mb-2">🏆 Game Over!</h1>

      {winners.length > 0 && (
        <div className="winner-banner mb-3">
          <span className="winner-label">{winners.length === 1 ? 'Winner' : 'Winners'}</span>
          <div className="winners-container">
            {winners.map((w) => (
              <div key={w.uuid} className="winner-name">
                <span className="player-avatar-large">{getAvatarForPlayerIndex(w.index)}</span>
                {w.nickname}
                <span className="winner-icon">👑</span>
              </div>
            ))}
          </div>
          <span className="winner-score">{winners[0].score.toLocaleString()} pts</span>
        </div>
      )}

      <div className="score-list mb-3">
        {sorted.map((p) => {
          const isWinner = p.score === maxScore && maxScore >= 0;
          const rank = sorted.findIndex((player) => player.score === p.score) + 1;
          return (
            <div key={p.uuid} className={`score-row ${p.uuid === state.myUUID ? 'me' : ''}`}>
              <span className="score-rank">#{rank}</span>
              <span className="score-name">
                <span className="player-avatar-small">{getAvatarForPlayerIndex(p.index)}</span>
                {p.nickname}
                {isWinner && <span className="winner-badge">👑</span>}
              </span>
              <span className="score-value">{p.score.toLocaleString()}</span>
            </div>
          );
        })}
      </div>

      {state.isHost && (
        <button className="btn-primary" onClick={handleRematch}>
          Play Again
        </button>
      )}
    </div>
  );
}
