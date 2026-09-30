import type { FromApiMapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { Project } from "@/model";

/**
 * API Project -> Project mapper. Projects are written through
 * toApiCreateProject and toApiUpdateProject.
 */
export const projectMapper: FromApiMapper<Project, Api.Project> = {
  fromApi(apiModel: Api.Project): Project {
    return {
      id: apiModel.id,
      name: apiModel.name,
      color: apiModel.color,
      timeBudgetHours: apiModel.timeBudgetHours,
      tagIds: new Set(apiModel.tagIds || []),
      tagId: apiModel.tagId,
      totalTimeMs: apiModel.totalTimeMs,
      ...(apiModel.taskTimeMs !== undefined && { taskTimeMs: apiModel.taskTimeMs }),
      archived: apiModel.archived,
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
    ...(patch.archived !== undefined && { archived: patch.archived }),
  };
}
