import { computed, readonly, ref } from "vue";

// Module-level singletons: startup is a whole-app concern, and the auth-redirect
// middleware (which runs outside any component) needs to flip the same flags the
// root component reads.
const probing = ref(true);
const redirecting = ref(false);

/**
 * Tracks the app's startup phase: the initial `/api/me` probe that decides
 * whether we have a session, and the subsequent bounce to the identity provider
 * when a data request comes back unauthorized. The UI shows a loading screen
 * while either is true so the user never sees empty chrome.
 */
export function useStartup() {
  const isStarting = computed(() => probing.value || redirecting.value);

  /** Call once the initial current-user probe has settled. */
  function finishProbe(): void {
    probing.value = false;
  }

  /** Call when navigation to the identity provider has been triggered. */
  function beginRedirect(): void {
    redirecting.value = true;
  }

  return {
    isStarting,
    redirecting: readonly(redirecting),
    finishProbe,
    beginRedirect,
  };
}

export const __test__ = {
  reset(): void {
    probing.value = true;
    redirecting.value = false;
  },
};
