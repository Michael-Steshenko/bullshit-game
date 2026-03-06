import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { GameProvider } from './context/GameContext';
import { Landing } from './pages/Landing';
import { CreateGame } from './pages/CreateGame';
import { JoinGame } from './pages/JoinGame';
import { GamePage } from './pages/GamePage';
import './styles/global.css';

export default function App() {
  return (
    <BrowserRouter>
      <GameProvider>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/create" element={<CreateGame />} />
          <Route path="/join" element={<JoinGame />} />
          <Route path="/game" element={<GamePage />} />
        </Routes>
      </GameProvider>
    </BrowserRouter>
  );
}
