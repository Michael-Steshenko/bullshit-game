import { useState, useRef } from 'react';
import { useGame } from '../hooks/useGame';
import { ProgressBar } from '../components/ProgressBar';
import './ShowQuestion.css';

export function ShowQuestion() {
  const { state, send } = useGame();
  const [answer, setAnswer] = useState('');
  const [submitted, setSubmitted] = useState(state.submittedPlayers.includes(state.myUUID));
  const timerExpiredRef = useRef(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!answer.trim()) return;
    send('submit_answer', {
      pin: state.pin,
      text: answer.trim(),
      stateVersion: state.stateVersion,
    });
    setSubmitted(true);
  };

  const handleExpired = () => {
    if (!timerExpiredRef.current) {
      timerExpiredRef.current = true;
      send('tick', { pin: state.pin, stateVersion: state.stateVersion });
    }
  };

  const questionText = state.question?.text.replace('$blank$', '______') || '';
  const showSubmitted = submitted && !state.error;

  return (
    <div className="show-question fade-in">
      <ProgressBar
        duration={state.duration}
        startTime={state.stateTimestamp}
        onExpired={handleExpired}
      />

      <div className="question-number mb-2">
        Question {state.question?.questionNumber} of {state.question?.totalQuestions}
      </div>

      <h2 className="question-text mb-3">{questionText}</h2>

      {showSubmitted ? (
        <div className="submitted-msg">
          <p>✅ Answer submitted!</p>
          <p className="submitted-count">
            {state.submittedPlayers.length} / {state.players.length} players answered
          </p>
        </div>
      ) : (
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            value={answer}
            onChange={(e) => setAnswer(e.target.value.slice(0, 40))}
            placeholder="Write a convincing lie..."
            maxLength={40}
            autoFocus
            className="answer-input"
          />
          <button className="btn-primary full-width mt-1" disabled={!answer.trim()}>
            Submit
          </button>
          {state.error && <p className="error-text">{state.error.message}</p>}
        </form>
      )}
    </div>
  );
}
