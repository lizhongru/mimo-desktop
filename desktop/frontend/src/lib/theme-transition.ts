/** Theme transition animation — circular clip-path reveal */

type Theme = "dark" | "light";

/** Calculate max radius from click point to farthest screen corner */
function maxRadius(x: number, y: number): number {
  const w = window.innerWidth;
  const h = window.innerHeight;
  return Math.hypot(Math.max(x, w - x), Math.max(y, h - y));
}

/** Apply theme to DOM + persist */
function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", theme === "dark");
  document.documentElement.classList.toggle("light", theme === "light");
  document.documentElement.style.colorScheme = theme;
  try { localStorage.setItem("mimo-theme", theme); } catch (_) {}
}

/** Path A: View Transitions API (Chromium 111+, Wails webview) */
function transitionWithVT(newTheme: Theme, x: number, y: number) {
  const root = document.documentElement;
  root.style.setProperty("--cx", x + "px");
  root.style.setProperty("--cy", y + "px");
  root.style.setProperty("--max-r", maxRadius(x, y) + "px");

  // @ts-ignore — startViewTransition exists in Chromium
  document.startViewTransition(() => {
    applyTheme(newTheme);
  });
}

/** Path B: rAF fallback — clip-path mask overlay */
function transitionWithFallback(newTheme: Theme, x: number, y: number) {
  const maxR = maxRadius(x, y);
  const duration = 450;
  const start = performance.now();

  // Create overlay with new theme background color
  const mask = document.createElement("div");
  mask.style.cssText = `
    position: fixed; inset: 0; z-index: 99999;
    pointer-events: none;
    clip-path: circle(0px at ${x}px ${y}px);
    transition: none;
  `;
  // Set overlay background to the new theme's bg color
  // We temporarily apply the new theme to a test element
  const isDark = newTheme === "dark";
  mask.style.background = isDark ? "#181818" : "#f8f8fa";
  document.body.appendChild(mask);

  function frame(now: number) {
    const t = Math.min((now - start) / duration, 1);
    // cubic easeOut: 1 - (1 - t)^3
    const ease = 1 - Math.pow(1 - t, 3);
    mask.style.clipPath = `circle(${ease * maxR}px at ${x}px ${y}px)`;

    if (t < 1) {
      requestAnimationFrame(frame);
    } else {
      applyTheme(newTheme);
      // Brief delay to let the real theme paint before removing mask
      requestAnimationFrame(() => mask.remove());
    }
  }
  requestAnimationFrame(frame);
}

/**
 * Animate theme switch with circular reveal.
 * @param newTheme target theme
 * @param originX click X coordinate (defaults to center)
 * @param originY click Y coordinate (defaults to center)
 */
export function animateThemeSwitch(
  newTheme: Theme,
  originX?: number,
  originY?: number
) {
  const current = document.documentElement.classList.contains("dark") ? "dark" : "light";
  if (newTheme === current) return;

  const x = originX ?? window.innerWidth / 2;
  const y = originY ?? window.innerHeight / 2;

  // Respect system reduced-motion preference
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    applyTheme(newTheme);
    return;
  }

  // Path A: View Transitions API
  if (typeof (document as any).startViewTransition === "function") {
    transitionWithVT(newTheme, x, y);
    return;
  }

  // Path B: fallback
  transitionWithFallback(newTheme, x, y);
}
