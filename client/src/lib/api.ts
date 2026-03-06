const API_BASE = import.meta.env.VITE_API_URL || '';

export async function createGame(lang: string, totalQuestions: number): Promise<{ pin: string }> {
  const res = await fetch(`${API_BASE}/api/create-game`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lang, totalQuestions }),
  });
  if (!res.ok) throw new Error('Failed to create game');
  return res.json();
}
