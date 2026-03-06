import { useState, useEffect, useRef } from 'react';

export function useTimer(duration: number, startTime: number) {
  const [remaining, setRemaining] = useState(duration);
  const rafRef = useRef<number>(0);

  useEffect(() => {
    if (duration <= 0) {
      setRemaining(0);
      return;
    }

    const update = () => {
      const elapsed = Date.now() - startTime;
      const left = Math.max(0, duration - elapsed);
      setRemaining(left);
      if (left > 0) {
        rafRef.current = requestAnimationFrame(update);
      }
    };

    rafRef.current = requestAnimationFrame(update);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [duration, startTime]);

  const seconds = Math.ceil(remaining / 1000);
  const progress = duration > 0 ? remaining / duration : 0;
  const isPanic = seconds <= 5 && seconds > 0;

  return { remaining, seconds, progress, isPanic };
}
