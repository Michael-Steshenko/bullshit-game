import { useGame } from '../hooks/useGame';
import { PlayerCard } from '../components/PlayerCard';
import './GameStaging.css';

export function GameStaging() {
  const { state, send } = useGame();

  const handleStart = () => {
    send('start_game', { pin: state.pin });
  };

  return (
    <div className="staging fade-in">
      <h2 className="mb-2">Lobby</h2>
      <div className="staging-pin mb-3">
        <span>Join at</span>
        <strong>{window.location.host}</strong>
        <span>with PIN</span>
        <strong className="pin-highlight">{state.pin}</strong>
      </div>

      <div className="player-grid mb-3">
        {state.players.map((p) => (
          <PlayerCard
            key={p.uuid}
            player={p}
            isHost={p.index === 0}
          />
        ))}
      </div>

      <p className="player-count mb-3">
        {state.players.length} / 8 players
      </p>

      {state.isHost && (
        <button className="btn-primary" onClick={handleStart}>
          Start Game
        </button>
      )}
      {!state.isHost && (
        <p className="waiting-text">Waiting for host to start...</p>
      )}
    </div>
  );
}
