import type { Task } from "@/model";
import {
  TasksApi as GeneratedTasksApi,
  GetTaskIncludeEnum,
  ListTasksIncludeEnum,
} from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { taskFromApi, toApiCreateTask, toApiUpdateTask, type TaskPatch } from "./mappers";
import { fetchAllPages } from "./pagination";

export type NewTask = Pick<Task, "name"> &
  Partial<Pick<Task, "parentId" | "tagIds" | "dueDate" | "estimateHours">>;

export interface TaskListFilter {
  includeClosed?: boolean;
  parentId?: string;
  /** Only subtasks of this task, at any depth. */
  ancestorId?: string;
  projectId?: string;
  /** Only tasks due on or after this date (YYYY-MM-DD). */
  dueFrom?: string;
  /** Only tasks due on or before this date (YYYY-MM-DD). */
  dueTo?: string;
}

export interface TasksApi {
  /**
   * Fetches every task matching filter, with its total time unless
   * withTotalTime is false.
   */
  listAllTasks(filter?: TaskListFilter, options?: { withTotalTime?: boolean }): Promise<Task[]>;
  /** Fetches one task; detailed adds its total time and projects. */
  getTask(id: string, detailed: boolean): Promise<Task>;
  createTask(task: NewTask): Promise<Task>;
  updateTask(id: string, patch: TaskPatch): Promise<Task>;
  /** Places the task under parentId, directly after afterTaskId (first if unset). */
  moveTask(id: string, parentId?: string, afterTaskId?: string): Promise<Task>;
  deleteTask(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedTasksApi(ApiConfig);

function createTasksApi(api: GeneratedTasksApi = defaultGeneratedApi): TasksApi {
  return {
    async listAllTasks(
      filter: TaskListFilter = {},
      { withTotalTime = true }: { withTotalTime?: boolean } = {},
    ): Promise<Task[]> {
      const { dueFrom, dueTo, ...rest } = filter;
      return fetchAllPages(async (limit, offset) => {
        const response = await api.listTasks({
          limit,
          offset,
          ...rest,
          ...(dueFrom && { dueFrom: new Date(dueFrom) }),
          ...(dueTo && { dueTo: new Date(dueTo) }),
          include: withTotalTime ? new Set([ListTasksIncludeEnum.TotalTimeMs]) : undefined,
        });
        return { data: response.data.map(taskFromApi), pagination: response.pagination };
      });
    },

    async getTask(id: string, detailed: boolean): Promise<Task> {
      const response = await api.getTask({
        taskId: id,
        include: detailed
          ? new Set([GetTaskIncludeEnum.TotalTimeMs, GetTaskIncludeEnum.ProjectIds])
          : undefined,
      });
      return taskFromApi(response);
    },

    async createTask(task: NewTask): Promise<Task> {
      const response = await api.createTask({ createTask: toApiCreateTask(task) });
      return taskFromApi(response);
    },

    async updateTask(id: string, patch: TaskPatch): Promise<Task> {
      const response = await api.updateTask({ taskId: id, updateTask: toApiUpdateTask(patch) });
      return taskFromApi(response);
    },

    async moveTask(id: string, parentId?: string, afterTaskId?: string): Promise<Task> {
      const response = await api.moveTask({
        taskId: id,
        moveTask: {
          ...(parentId && { parentId }),
          ...(afterTaskId && { afterTaskId }),
        },
      });
      return taskFromApi(response);
    },

    async deleteTask(id: string): Promise<void> {
      await api.deleteTask({ taskId: id });
    },
  };
}

export const tasksApi = createTasksApi();
export const __test__ = { createTasksApi };
