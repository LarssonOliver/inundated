import { computed, ref } from "vue";

// Module-level singletons: startup is a whole-app concern, and the auth-redirect
// middleware (which runs outside any component) needs to flip the same flags the
// root component reads. The mutators and readers below are plain exports so
// non-component callers don't reconstruct a return object on every use.
const probing = ref(true);
const redirecting = ref(false);

/** True while either the initial probe or the IdP redirect is in progress. */
const isStarting = computed(() => probing.value || redirecting.value);

/** Call once the initial current-user probe has settled (or timed out). */
export function finishProbe(): void {
  probing.value = false;
}

/** Call when navigation to the identity provider has been triggered. */
export function beginRedirect(): void {
  redirecting.value = true;
}

/** Whether a redirect to the identity provider is already under way. */
export function isRedirecting(): boolean {
  return redirecting.value;
}

/**
 * How long the loading screen waits on the startup probe before giving up and
 * rendering the app anyway.
 */
export const PROBE_TIMEOUT_MS = 8000;

/**
 * Awaits the initial current-user probe, then clears the probing flag. Stops
 * blocking after {@link PROBE_TIMEOUT_MS} so a hung `/api/me` request can't
 * strand the user on the spinner; the probe keeps running and the user store
 * still updates if it resolves later. Probe errors are left for the pinia error
 * plugin to record.
 */
export async function runStartupProbe(probe: Promise<unknown>): Promise<void> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<void>((resolve) => {
    timer = setTimeout(resolve, PROBE_TIMEOUT_MS);
  });
  try {
    await Promise.race([probe.catch(() => undefined), timeout]);
  } finally {
    clearTimeout(timer);
    finishProbe();
  }
}

/**
 * Tracks the app's startup phase: the initial `/api/me` probe that decides
 * whether we have a session, and the subsequent bounce to the identity provider
 * when a data request comes back unauthorized. The UI shows a loading screen
 * while either is true so the user never sees empty chrome.
 */
export function useStartup() {
  return { isStarting, finishProbe, beginRedirect };
}

export const __test__ = {
  reset(): void {
    probing.value = true;
    redirecting.value = false;
  },
};
