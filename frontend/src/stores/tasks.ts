import { tasksApi, type NewTask, type TasksApi } from "@/api";
import type { TaskPatch } from "@/api/mappers";
import { useSupersededFetch } from "@/composables/useSupersededFetch";
import type { Task } from "@/model";
import { acceptHMRUpdate, defineStore } from "pinia";
import { computed, ref } from "vue";

function copyTask(task: Task): Task {
  return {
    ...task,
    tagIds: new Set(task.tagIds),
    projectIds: task.projectIds ? new Set(task.projectIds) : undefined,
  };
}

/** A task placed in the tree: depth 0 for top-level tasks. */
export interface TaskRow {
  task: Task;
  depth: number;
}

/**
 * Orders tasks as a depth-first tree, each parent followed by its subtasks,
 * siblings by rank. A task whose parent isn't in the list is shown at the
 * top level, so a filtered list never hides it.
 */
export function taskTree(tasks: readonly Task[]): TaskRow[] {
  const ids = new Set(tasks.map((task) => task.id));
  const children = new Map<string | undefined, Task[]>();
  for (const task of tasks) {
    const parent = task.parentId && ids.has(task.parentId) ? task.parentId : undefined;
    const siblings = children.get(parent) ?? [];
    siblings.push(task);
    children.set(parent, siblings);
  }
  for (const siblings of children.values()) {
    siblings.sort((a, b) => (a.rank < b.rank ? -1 : a.rank > b.rank ? 1 : 0));
  }

  const rows: TaskRow[] = [];
  const visit = (parent: string | undefined, depth: number) => {
    for (const task of children.get(parent) ?? []) {
      rows.push({ task, depth });
      visit(task.id, depth + 1);
    }
  };
  visit(undefined, 0);
  return rows;
}

function createTasksStore(api: TasksApi) {
  return defineStore("tasks", () => {
    const tasks = ref<Map<string, Task>>(new Map());
    // Tasks fetched individually by id, e.g. a closed task opened from a
    // link while closed tasks are hidden. Kept apart from `tasks` so they
    // don't leak into the list.
    const individuallyFetchedTasks = ref<Map<string, Task>>(new Map());
    const includeClosed = ref(false);
    const supersededFetch = useSupersededFetch();
    const isLoading = supersededFetch.isLoading;

    const readOnlyTasks = computed<readonly Task[]>(() =>
      Array.from(tasks.value.values()).map(copyTask),
    );

    /**
     * Fetches every task (closed ones too when includeClosed is set), with
     * its total time, replacing the local list.
     */
    async function fetchTasks(): Promise<void> {
      const key = `all:${includeClosed.value}`;
      await supersededFetch.run(key, async () => {
        const result = await api.listAllTasks({ includeClosed: includeClosed.value });
        if (supersededFetch.isStale(key)) return;
        tasks.value = new Map(result.map((task) => [task.id, task]));
      });
    }

    /**
     * Sets whether closed tasks are listed, reloading the list if it changes.
     */
    async function setIncludeClosed(value: boolean): Promise<void> {
      if (includeClosed.value === value) return;
      includeClosed.value = value;
      await fetchTasks();
    }

    function getTaskById(id: string): Task | undefined {
      const task = tasks.value.get(id) ?? individuallyFetchedTasks.value.get(id);
      return task ? copyTask(task) : undefined;
    }

    /**
     * Fetches a task with its total time and projects, bypassing the list.
     */
    async function fetchDetailedTaskById(id: string): Promise<Task> {
      const task = await api.getTask(id, true);
      individuallyFetchedTasks.value.set(id, task);
      return copyTask(task);
    }

    async function createTask(task: NewTask): Promise<Task> {
      const created = await api.createTask(task);
      tasks.value.set(created.id, created);
      return copyTask(created);
    }

    /**
     * Creates a task from a name, as typed after "#" in the tag picker.
     */
    async function createTaskFromName(name: string, parentId?: string): Promise<Task> {
      const normalizedName = name.trim();
      if (!normalizedName) {
        throw new Error("Task name cannot be empty");
      }
      return await createTask({ name: normalizedName, parentId });
    }

    /**
     * Updates a task. Closing or reopening one also changes its subtasks or
     * parents on the server, so the list is reloaded afterwards.
     */
    async function updateTask(id: string, patch: TaskPatch): Promise<Task> {
      const updated = await api.updateTask(id, patch);
      individuallyFetchedTasks.value.delete(id);
      if (patch.closed !== undefined || patch.closeReason !== undefined) {
        await fetchTasks();
      } else if (tasks.value.has(id)) {
        tasks.value.set(id, { ...updated, totalTimeMs: tasks.value.get(id)?.totalTimeMs });
      }
      return copyTask(updated);
    }

    async function closeTask(id: string, reason: Task["closeReason"] = "done"): Promise<Task> {
      return await updateTask(id, { closed: true, closeReason: reason });
    }

    async function reopenTask(id: string): Promise<Task> {
      return await updateTask(id, { closed: false });
    }

    /**
     * Moves a task under parentId (top level when unset), directly after
     * afterTaskId (first when unset). Siblings may be re-ranked, so the list
     * is reloaded afterwards.
     */
    async function moveTask(id: string, parentId?: string, afterTaskId?: string): Promise<Task> {
      const moved = await api.moveTask(id, parentId, afterTaskId);
      await fetchTasks();
      return copyTask(moved);
    }

    /**
     * Moves a task one step up or down among its siblings in the list.
     */
    async function shiftTask(id: string, delta: -1 | 1): Promise<void> {
      const task = tasks.value.get(id);
      if (!task) return;
      const siblings = taskTree(readOnlyTasks.value)
        .map((row) => row.task)
        .filter((t) => t.parentId === task.parentId);
      const index = siblings.findIndex((t) => t.id === id);
      const target = index + delta;
      if (index === -1 || target < 0 || target >= siblings.length) return;

      // Moving up places it after the sibling two above (or first); moving
      // down places it after the next sibling.
      const others = siblings.filter((t) => t.id !== id);
      const after = target === 0 ? undefined : others[target - 1];
      await moveTask(id, task.parentId, after?.id);
    }

    /**
     * Deletes a task and its subtasks. The server refuses (409) when any of
     * them has logged time.
     */
    async function deleteTask(id: string): Promise<void> {
      await api.deleteTask(id);
      individuallyFetchedTasks.value.delete(id);
      await fetchTasks();
    }

    return {
      tasks: readOnlyTasks,
      isLoading,
      includeClosed,
      setIncludeClosed,
      fetchTasks,
      getTaskById,
      fetchDetailedTaskById,
      createTask,
      createTaskFromName,
      updateTask,
      closeTask,
      reopenTask,
      moveTask,
      shiftTask,
      deleteTask,
    };
  });
}

export const useTasksStore = createTasksStore(tasksApi);
export const __test__ = { createTasksStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTasksStore, import.meta.hot));
}
