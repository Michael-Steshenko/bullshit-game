import { useState, useEffect, useRef } from 'react';

export function useTimer(duration: number, startTime: number) {
  const [remaining, setRemaining] = useState(() => {
    const initialElapsed = Date.now() - startTime;
    return Math.max(0, duration - initialElapsed);
  });
  const rafRef = useRef<number>(0);

  useEffect(() => {
    const update = () => {
      const elapsed = Date.now() - startTime;
      const left = Math.max(0, duration - elapsed);
      setRemaining(left);
      if (left > 0) {
        rafRef.current = requestAnimationFrame(update);
      }
    };

    if (duration > 0) {
      rafRef.current = requestAnimationFrame(update);
    } else {
      rafRef.current = requestAnimationFrame(() => setRemaining(0));
    }

    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [duration, startTime]);

  const seconds = Math.ceil(remaining / 1000);
  const progress = duration > 0 ? remaining / duration : 0;
  const isPanic = seconds <= 5 && seconds > 0;

  return { remaining, seconds, progress, isPanic };
}
