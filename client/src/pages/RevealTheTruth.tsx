import { useState, useEffect } from 'react';
import { useGame } from '../hooks/useGame';
import { getAvatarForPlayerIndex } from '../lib/avatar';
import './RevealTheTruth.css';

export function RevealTheTruth() {
  const { state, send } = useGame();
  const [currentIdx, setCurrentIdx] = useState(0);
  const [phase, setPhase] = useState<'show' | 'reveal'>('show');

  const reveals = state.reveals;

  useEffect(() => {
    if (reveals.length === 0) return;

    // Phase timing: show answer (3s), reveal creator (4s), then next
    const showTimer = setTimeout(() => {
      setPhase('reveal');
    }, 3000);

    const nextTimer = setTimeout(() => {
      if (currentIdx < reveals.length - 1) {
        setCurrentIdx((i) => i + 1);
        setPhase('show');
      } else {
        // All revealed, advance
        send('tick', { pin: state.pin, stateVersion: state.stateVersion });
      }
    }, 7000);

    return () => {
      clearTimeout(showTimer);
      clearTimeout(nextTimer);
    };
  }, [currentIdx, reveals.length, send, state.pin, state.stateVersion]);

  if (reveals.length === 0) return <div>Loading...</div>;

  const current = reveals[currentIdx];
  const players = state.players;
  const hasWrittenBy = !current.realAnswer && !(current.houseLie && current.creators[0] === 'house');

  const getPlayer = (uuid: string) => players.find((p) => p.uuid === uuid);

  const getPlayerLabel = (uuid: string) => {
    if (uuid === 'house') return '🏠 Home Grown Bullshit';
    if (uuid === 'truth') return '✅ The Truth';
    return getPlayer(uuid)?.nickname || uuid;
  };

  const renderPlayerChip = (uuid: string) => {
    if (uuid === 'house' || uuid === 'truth') {
      return (
        <span key={uuid} className="reveal-person-chip system">
          <span>{getPlayerLabel(uuid)}</span>
        </span>
      );
    }

    const player = getPlayer(uuid);
    if (!player) {
      return (
        <span key={uuid} className="reveal-person-chip">
          <span>{uuid}</span>
        </span>
      );
    }

    return (
      <span key={uuid} className="reveal-person-chip">
        <span className="reveal-person-avatar">{getAvatarForPlayerIndex(player.index)}</span>
        <span>{player.nickname}</span>
      </span>
    );
  };

  return (
    <div className="reveal fade-in">
      <div className="reveal-progress">
        {currentIdx + 1} / {reveals.length}
      </div>

      <div className="reveal-layout">
        <div className="reveal-writers">
          {hasWrittenBy && <span className="reveal-label">Written by</span>}
          <div className="reveal-people-row">
            {current.creators.map((uuid) => renderPlayerChip(uuid))}
          </div>
          {phase === 'reveal' && current.creatorPoints > 0 && (
            <div className="reveal-points positive">+{current.creatorPoints} </div>
          )}
        </div>

        <div
          className={`reveal-card ${current.realAnswer ? 'real' : current.houseLie ? 'house' : 'player'}`}
        >
          <h2 className="reveal-text">{current.text}</h2>
        </div>

        {current.selectors?.length > 0 && (
          <div className="reveal-selectors">
            <span className="reveal-label">Selected by</span>
            <div className="reveal-people-row">
              {current.selectors.map((uuid) => renderPlayerChip(uuid))}
            </div>
            {phase === 'reveal' && current.selectorPoints !== 0 && (
              <div
                className={`reveal-points ${current.selectorPoints > 0 ? 'positive' : 'negative'}`}
              >
                {current.selectorPoints > 0 ? '+' : ''}
                {current.selectorPoints}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
