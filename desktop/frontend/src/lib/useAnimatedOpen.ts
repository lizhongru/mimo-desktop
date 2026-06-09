import { useState, useEffect, useRef } from "react";

/**
 * Manages open/close state with exit animation delay.
 * Returns { isOpen, shouldRender, closing }.
 * - shouldRender: true while the element should be in the DOM
 * - closing: true during the exit animation
 *
 * Usage:
 *   const { shouldRender, closing } = useAnimatedOpen(rawOpen, 150);
 *   {shouldRender && <div className={closing ? "animate-pop-out" : "animate-pop-up"}>...</div>}
 */
export function useAnimatedOpen(open: boolean, duration = 150) {
  const [shouldRender, setShouldRender] = useState(open);
  const [closing, setClosing] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);
  const prevOpen = useRef(open);

  useEffect(() => {
    clearTimeout(timer.current);

    if (open && !prevOpen.current) {
      setClosing(false);
      setShouldRender(true);
    } else if (!open && prevOpen.current) {
      setClosing(true);
      timer.current = setTimeout(() => {
        setClosing(false);
        setShouldRender(false);
      }, duration);
    }

    prevOpen.current = open;
    return () => { clearTimeout(timer.current); };
  }, [open, duration]);

  return { shouldRender, closing };
}
