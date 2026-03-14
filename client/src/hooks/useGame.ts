import { useContext } from 'react';
import { GameContext, type GameContextType } from '../context/GameContext.types';

export function useGame(): GameContextType {
  const ctx = useContext(GameContext);
  if (!ctx) throw new Error('useGame must be used within GameProvider');
  return ctx;
}
