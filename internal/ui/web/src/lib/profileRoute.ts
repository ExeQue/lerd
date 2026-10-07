import { get } from 'svelte/store';
import { profilerEnabled, setProfiler, captureCount, waitForCapture } from '$stores/profiler';

export type ProfilePhase = 'arming' | 'waiting';

// profileRoute arms the profiler, opens the route, and reports whether SPX
// caught it. Each step waits for the one before it: arming only returns once
// nginx serves the profiling config, so the request cannot be answered by the
// configuration with no profiler attached, and the capture count only rises
// once the report is on disk. A profiler this armed is put back after, rather
// than leaving every FPM site profiled.
export async function profileRoute(host: string, route: string, url: string, onPhase: (p: ProfilePhase) => void): Promise<boolean> {
  const armedHere = !get(profilerEnabled);
  try {
    const before = await captureCount(host, route);
    if (armedHere) {
      onPhase('arming');
      await setProfiler(true);
    }
    // Opened once, here, with the real URL. Holding a blank tab open across the
    // arming wait leaves an about:blank the desktop is asked to find an
    // application for when the dashboard runs as an app window.
    window.open(url, '_blank');
    onPhase('waiting');
    return await waitForCapture(host, route, before);
  } catch {
    return false;
  } finally {
    if (armedHere) {
      try {
        await setProfiler(false);
      } catch {
        /* it stays armed; the toggle is one click away */
      }
    }
  }
}
