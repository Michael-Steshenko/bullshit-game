import { useCallback } from 'react';

const SESSION_KEY = 'bullshit_session';

interface SessionData {
  pin: string;
  uuid: string;
  nickname: string;
}

export function useSession() {
  const getSession = useCallback((): SessionData | null => {
    try {
      const raw = sessionStorage.getItem(SESSION_KEY);
      if (!raw) return null;
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }, []);

  const setSession = useCallback((data: SessionData) => {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(data));
  }, []);

  const clearSession = useCallback(() => {
    sessionStorage.removeItem(SESSION_KEY);
  }, []);

  return { getSession, setSession, clearSession };
}
