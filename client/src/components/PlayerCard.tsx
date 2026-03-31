import { getAvatarForPlayerIndex } from '../lib/avatar';
import type { PlayerData } from '../lib/types';
import './PlayerCard.css';

interface Props {
  player: PlayerData;
  isHost?: boolean;
  showScore?: boolean;
}

export function PlayerCard({ player, isHost, showScore }: Props) {
  const avatar = getAvatarForPlayerIndex(player.index);

  return (
    <div className="player-card fade-in">
      <div className="player-avatar">{avatar}</div>
      <div className="player-info">
        <span className="player-name">
          {player.nickname}
          {isHost && <span className="host-badge">👑</span>}
        </span>
        {showScore && <span className="player-score">{player.score.toLocaleString()}</span>}
      </div>
    </div>
  );
}
