import type { User } from "@/model";
import {
  AuthApi as GeneratedAuthApi,
  ResponseError,
  UsersApi as GeneratedUsersApi,
} from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { xsrfToken } from "@/api/xsrf";

export interface UsersApi {
  /**
   * Returns the signed-in user, or null when there is no session. A null result
   * is expected both when auth is enabled and the visitor is logged out and when
   * the server runs in userless mode.
   */
  getCurrentUser(): Promise<User | null>;
  logout(): Promise<void>;
}

function createUsersApi(
  users: GeneratedUsersApi = new GeneratedUsersApi(ApiConfig),
  auth: GeneratedAuthApi = new GeneratedAuthApi(ApiConfig),
): UsersApi {
  return {
    async getCurrentUser(): Promise<User | null> {
      try {
        const user = await users.getCurrentUser();
        return { id: user.id, email: user.email, name: user.name ?? "" };
      } catch (error) {
        if (error instanceof ResponseError && error.response.status === 401) {
          return null;
        }
        throw error;
      }
    },

    async logout(): Promise<void> {
      try {
        await auth.authLogout({ xXSRFTOKEN: xsrfToken() });
      } catch (error) {
        // A 401 here means the session was already gone server-side; that is
        // the outcome logout wanted, so let the caller proceed to reload.
        if (error instanceof ResponseError && error.response.status === 401) {
          return;
        }
        throw error;
      }
    },
  };
}

export const usersApi = createUsersApi();
export const __test__ = { createUsersApi };
