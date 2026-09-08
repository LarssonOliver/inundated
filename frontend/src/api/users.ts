import type { User } from "@/model";
import {
  AuthApi as GeneratedAuthApi,
  ResponseError,
  UsersApi as GeneratedUsersApi,
} from "@/api/generated";
import { ApiConfig } from "@/api/config";

export interface UsersApi {
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
        await auth.authLogout();
      } catch (error) {
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
