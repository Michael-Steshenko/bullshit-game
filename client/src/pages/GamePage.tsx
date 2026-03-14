import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useGame } from '../hooks/useGame';
import { GameState } from '../lib/types';
import { GameHeader } from '../components/GameHeader';
import { GameFooter } from '../components/GameFooter';
import { GameStaging } from './GameStaging';
import { RoundIntro } from './RoundIntro';
import { ShowQuestion } from './ShowQuestion';
import { ShowAnswers } from './ShowAnswers';
import { RevealTheTruth } from './RevealTheTruth';
import { ScoreBoard } from './ScoreBoard';
import { ScoreBoardFinal } from './ScoreBoardFinal';

export function GamePage() {
  const { state } = useGame();
  const navigate = useNavigate();

  // Redirect if not in a game
  useEffect(() => {
    if (!state.myUUID) {
      navigate('/');
    }
  }, [state.myUUID, navigate]);

  const renderState = () => {
    switch (state.gameState) {
      case GameState.GameStaging:
        return <GameStaging />;
      case GameState.RoundIntro:
        return <RoundIntro />;
      case GameState.ShowQuestion:
        return <ShowQuestion />;
      case GameState.ShowAnswers:
        return <ShowAnswers />;
      case GameState.RevealTheTruth:
        return <RevealTheTruth />;
      case GameState.ScoreBoard:
        return <ScoreBoard />;
      case GameState.ScoreBoardFinal:
        return <ScoreBoardFinal />;
      default:
        return <div>Unknown state</div>;
    }
  };

  return (
    <>
      <GameHeader />
      <main className="page" style={{ flex: 1 }}>
        {renderState()}
      </main>
      <GameFooter />
    </>
  );
}
