import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { Project } from "@/model";

/**
 * Full Project <-> API Project mapper
 */
export const projectMapper: Mapper<Project, Api.Project> = {
  fromApi(apiModel: Api.Project): Project {
    return {
      id: apiModel.id,
      name: apiModel.name,
      color: apiModel.color,
      timeBudgetHours: apiModel.timeBudgetHours,
      tagIds: new Set(apiModel.tagIds || []),
    };
  },
  toApi(domainModel: Project): Api.Project {
    return {
      id: domainModel.id,
      name: domainModel.name,
      color: domainModel.color,
      timeBudgetHours: domainModel.timeBudgetHours,
      tagIds: domainModel.tagIds.size > 0 ? new Set(domainModel.tagIds) : undefined,
    };
  },
};

/**
 * CreateProject mapper (domain -> API)
 */
export function toApiCreateProject(domain: Omit<Project, "id">): Api.CreateProject {
  return {
    name: domain.name,
    color: domain.color,
    timeBudgetHours: domain.timeBudgetHours,
    tagIds: new Set(domain.tagIds),
  };
}

/**
 * UpdateProject mapper (partial domain -> API)
 */
export function toApiUpdateProject(patch: Partial<Omit<Project, "id">>): Api.UpdateProject {
  return {
    ...(patch.name !== undefined && { name: patch.name }),
    ...(patch.color !== undefined && { color: patch.color }),
    ...(patch.timeBudgetHours !== undefined && { timeBudgetHours: patch.timeBudgetHours }),
    ...(patch.tagIds !== undefined && { tagIds: new Set(patch.tagIds) }),
  };
}
