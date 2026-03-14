import { useGame } from '../hooks/useGame';
import { ProgressBar } from '../components/ProgressBar';
import './ScoreBoard.css';

export function ScoreBoard() {
  const { state, send } = useGame();

  const handleExpired = () => {
    send('tick', { pin: state.pin, stateVersion: state.stateVersion });
  };

  const sorted = [...state.players].sort((a, b) => b.score - a.score);

  return (
    <div className="scoreboard fade-in">
      <ProgressBar
        duration={state.duration}
        startTime={state.stateTimestamp}
        onExpired={handleExpired}
      />

      <h2 className="mb-3">Scores</h2>

      <div className="score-list">
        {sorted.map((p, i) => (
          <div key={p.uuid} className={`score-row ${p.uuid === state.myUUID ? 'me' : ''}`}>
            <span className="score-rank">#{i + 1}</span>
            <span className="score-name">{p.nickname}</span>
            <span className="score-value">{p.score.toLocaleString()}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
