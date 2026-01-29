import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { Tag } from "@/model";

/**
 * Full Tag <-> API Tag mapper
 */
export const tagMapper: Mapper<Tag, Api.Tag> = {
  fromApi(apiModel: Api.Tag): Tag {
    return {
      id: apiModel.id,
      name: apiModel.name,
      color: apiModel.color,
    };
  },
  toApi(domainModel: Tag): Api.Tag {
    return {
      id: domainModel.id,
      name: domainModel.name,
      color: domainModel.color,
    };
  },
};

/**
 * CreateTag mapper (domain -> API)
 */
export function toApiCreateTag(domain: Omit<Tag, "id">): Api.CreateTag {
  return {
    name: domain.name,
    color: domain.color,
  };
}

/**
 * UpdateTag mapper (partial domain -> API)
 */
export function toApiUpdateTag(patch: Partial<Omit<Tag, "id">>): Api.UpdateTag {
  return {
    ...(patch.name !== undefined && { name: patch.name }),
    ...(patch.color !== undefined && { color: patch.color }),
  };
}
