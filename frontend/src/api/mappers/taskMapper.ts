import type * as Api from "@/api/generated/models";
import type { Task } from "@/model";

/** Converts an API date (midnight UTC) to a YYYY-MM-DD day. */
function toDay(date: Date): string {
  return date.toISOString().slice(0, 10);
}

/** Converts a YYYY-MM-DD day to the API's midnight-UTC date. */
function fromDay(day: string): Date {
  return new Date(`${day}T00:00:00Z`);
}

export function taskFromApi(apiModel: Api.Task): Task {
  return {
    id: apiModel.id,
    name: apiModel.name,
    tagId: apiModel.tagId,
    parentId: apiModel.parentId,
    tagIds: new Set(apiModel.tagIds || []),
    dueDate: apiModel.dueDate ? toDay(apiModel.dueDate) : undefined,
    estimateHours: apiModel.estimateHours,
    rank: apiModel.rank,
    closed: apiModel.closed,
    closeReason: apiModel.closeReason,
    closedAt: apiModel.closedAt,
    totalTimeMs: apiModel.totalTimeMs,
    projectIds: apiModel.projectIds ? new Set(apiModel.projectIds) : undefined,
  };
}

/**
 * CreateTask mapper (domain -> API)
 */
export function toApiCreateTask(
  domain: Pick<Task, "name"> &
    Partial<Pick<Task, "parentId" | "tagIds" | "dueDate" | "estimateHours">>,
): Api.CreateTask {
  return {
    name: domain.name,
    ...(domain.parentId && { parentId: domain.parentId }),
    ...(domain.tagIds && domain.tagIds.size > 0 && { tagIds: new Set(domain.tagIds) }),
    ...(domain.dueDate && { dueDate: fromDay(domain.dueDate) }),
    ...(domain.estimateHours !== undefined && { estimateHours: domain.estimateHours }),
  };
}

/**
 * The fields of a task an update can change. A due date or estimate set to
 * null is cleared; one left undefined is kept.
 */
export interface TaskPatch {
  name?: string;
  tagIds?: Set<string>;
  dueDate?: string | null;
  estimateHours?: number | null;
  closed?: boolean;
  closeReason?: Task["closeReason"];
}

/**
 * UpdateTask mapper (partial domain -> API)
 */
export function toApiUpdateTask(patch: TaskPatch): Api.UpdateTask {
  return {
    ...(patch.name !== undefined && { name: patch.name }),
    ...(patch.tagIds !== undefined && { tagIds: new Set(patch.tagIds) }),
    ...(patch.dueDate === null && { clearDueDate: true }),
    ...(patch.dueDate != null && { dueDate: fromDay(patch.dueDate) }),
    ...(patch.estimateHours === null && { clearEstimate: true }),
    ...(patch.estimateHours != null && { estimateHours: patch.estimateHours }),
    ...(patch.closed !== undefined && { closed: patch.closed }),
    ...(patch.closeReason !== undefined && { closeReason: patch.closeReason }),
  };
}
