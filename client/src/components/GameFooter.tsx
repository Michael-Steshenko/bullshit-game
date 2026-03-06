import { useGame } from '../context/GameContext';
import './GameFooter.css';

export function GameFooter() {
  const { state } = useGame();

  if (!state.myNickname) return null;

  return (
    <footer className="game-footer">
      <span className="footer-nickname">{state.myNickname}</span>
      <span className="footer-score">{state.myScore.toLocaleString()} pts</span>
    </footer>
  );
}
