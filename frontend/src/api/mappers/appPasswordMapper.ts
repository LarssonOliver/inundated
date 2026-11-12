import type { FromApiMapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { AppPassword } from "@/model";

export const appPasswordMapper: FromApiMapper<AppPassword, Api.AppPassword> = {
  fromApi(apiModel: Api.AppPassword): AppPassword {
    return {
      id: apiModel.id,
      name: apiModel.name,
      createdAt: apiModel.createdAt,
      lastUsedAt: apiModel.lastUsedAt ?? null,
    };
  },
};
