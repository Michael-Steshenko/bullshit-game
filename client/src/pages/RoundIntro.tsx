import { useGame } from '../context/GameContext';
import { ProgressBar } from '../components/ProgressBar';
import './RoundIntro.css';

const ROUND_NAMES = ['Round 1', 'Round 2', 'Final Round'];

export function RoundIntro() {
  const { state, send } = useGame();

  const handleExpired = () => {
    send('tick', { pin: state.pin, stateVersion: state.stateVersion });
  };

  return (
    <div className="round-intro fade-in">
      <ProgressBar
        duration={state.duration}
        startTime={state.stateTimestamp}
        onExpired={handleExpired}
      />
      <h1 className="round-name pulse">
        {ROUND_NAMES[state.roundIndex] || `Round ${state.roundIndex + 1}`}
      </h1>
      <p className="round-subtitle">
        Question {state.questionIndex + 1} of {state.totalQuestions}
      </p>
    </div>
  );
}
