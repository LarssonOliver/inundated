import { usersApi, type UsersApi } from "@/api/users";
import type { User } from "@/model";
import { acceptHMRUpdate, defineStore } from "pinia";
import { computed, ref } from "vue";

function createUserStore(api: UsersApi) {
  return defineStore("user", () => {
    const user = ref<User | null>(null);

    const isAuthenticated = computed(() => user.value !== null);

    /**
     * Loads the signed-in user. Leaves the store unauthenticated when there is
     * no session, which is also the steady state in userless mode.
     */
    async function fetchCurrentUser(): Promise<void> {
      user.value = await api.getCurrentUser();
    }

    /**
     * Ends the session and reloads at the root, where the app re-probes auth.
     */
    async function logout(): Promise<void> {
      await api.logout();
      window.location.assign("/");
    }

    return { user, isAuthenticated, fetchCurrentUser, logout };
  });
}

export const useUserStore = createUserStore(usersApi);
export const __test__ = { createUserStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useUserStore, import.meta.hot));
}
