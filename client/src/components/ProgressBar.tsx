import { useTimer } from '../hooks/useTimer';
import './ProgressBar.css';

interface Props {
  duration: number;
  startTime: number;
  onExpired: () => void;
}

export function ProgressBar({ duration, startTime, onExpired }: Props) {
  const { seconds, progress, isPanic } = useTimer(duration, startTime);

  if (duration <= 0) return null;

  if (seconds === 0) {
    // Fire onExpired once
    setTimeout(onExpired, 0);
  }

  return (
    <div className={`progress-bar-container ${isPanic ? 'panic' : ''}`}>
      <div className="progress-bar-fill" style={{ width: `${progress * 100}%` }} />
      <span className="progress-bar-text">{seconds}s</span>
    </div>
  );
}
