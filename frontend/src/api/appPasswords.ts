import type { AppPassword } from "@/model";
import { AppPasswordsApi as GeneratedAppPasswordsApi } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { appPasswordMapper } from "./mappers";

export interface AppPasswordsApi {
  list(): Promise<AppPassword[]>;
  /** The plaintext password is only ever returned here. */
  create(name: string): Promise<{ appPassword: AppPassword; password: string }>;
  delete(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedAppPasswordsApi(ApiConfig);

function createAppPasswordsApi(
  api: GeneratedAppPasswordsApi = defaultGeneratedApi,
): AppPasswordsApi {
  return {
    async list(): Promise<AppPassword[]> {
      const response = await api.listAppPasswords();
      return response.map(appPasswordMapper.fromApi);
    },

    async create(name: string) {
      const { password, ...created } = await api.createAppPassword({ createAppPassword: { name } });
      return { appPassword: appPasswordMapper.fromApi(created), password };
    },

    async delete(id: string): Promise<void> {
      await api.deleteAppPassword({ appPasswordId: id });
    },
  };
}

export const appPasswordsApi = createAppPasswordsApi();
export const __test__ = { createAppPasswordsApi };
