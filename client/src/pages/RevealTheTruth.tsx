import { useState, useEffect } from 'react';
import { useGame } from '../hooks/useGame';
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
        setCurrentIdx(i => i + 1);
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

  const getPlayerName = (uuid: string) => {
    if (uuid === 'house') return '🏠 House Lie';
    if (uuid === 'truth') return '✅ The Truth';
    return players.find(p => p.uuid === uuid)?.nickname || uuid;
  };

  return (
    <div className="reveal fade-in">
      <div className="reveal-progress">
        {currentIdx + 1} / {reveals.length}
      </div>

      <div className={`reveal-card ${current.realAnswer ? 'real' : current.houseLie ? 'house' : 'player'}`}>
        <h2 className="reveal-text">{current.text}</h2>

        {current.selectors && current.selectors.length > 0 && (
          <div className="reveal-selectors">
            <span className="reveal-label">Selected by:</span>
            <div className="selector-names">
              {current.selectors.map(uuid => (
                <span key={uuid} className="selector-name">{getPlayerName(uuid)}</span>
              ))}
            </div>
          </div>
        )}

        {phase === 'reveal' && (
          <div className="reveal-creator fade-in">
            <span className="reveal-label">Written by:</span>
            <div className="creator-names">
              {current.creators.map(uuid => (
                <span key={uuid} className="creator-name">{getPlayerName(uuid)}</span>
              ))}
            </div>
            {current.points !== 0 && (
              <div className={`reveal-points ${current.points > 0 ? 'positive' : 'negative'}`}>
                {current.points > 0 ? '+' : ''}{current.points}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
