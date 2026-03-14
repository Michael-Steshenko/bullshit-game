import { useState, useRef } from 'react';
import { useGame } from '../hooks/useGame';
import { ProgressBar } from '../components/ProgressBar';
import './ShowAnswers.css';

export function ShowAnswers() {
  const { state, send } = useGame();
  const [selected, setSelected] = useState<string | null>(null);
  const timerExpiredRef = useRef(false);

  const handleSelect = (text: string) => {
    if (selected) return;
    setSelected(text);
    send('select_answer', {
      pin: state.pin,
      text,
      stateVersion: state.stateVersion,
    });
  };

  const handleExpired = () => {
    if (!timerExpiredRef.current) {
      timerExpiredRef.current = true;
      send('tick', { pin: state.pin, stateVersion: state.stateVersion });
    }
  };

  return (
    <div className="show-answers fade-in">
      <ProgressBar
        duration={state.duration}
        startTime={state.stateTimestamp}
        onExpired={handleExpired}
      />

      <h2 className="mb-2">Pick the truth!</h2>

      <p className="select-count mb-3">
        {state.selectedPlayers.length} / {state.players.length} voted
      </p>

      <div className="answer-grid">
        {state.answers.map((a, i) => (
          <button
            key={i}
            className={`answer-option ${selected === a.text ? 'selected' : ''} ${selected && selected !== a.text ? 'dimmed' : ''}`}
            onClick={() => handleSelect(a.text)}
            disabled={!!selected}
          >
            {a.text}
          </button>
        ))}
      </div>
    </div>
  );
}
