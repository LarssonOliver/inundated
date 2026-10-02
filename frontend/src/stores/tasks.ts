import { tasksApi, type NewTask, type TasksApi } from "@/api";
import type { TaskPatch } from "@/api/mappers";
import { useSupersededFetch } from "@/composables/useSupersededFetch";
import { createSerialQueue } from "@/helpers/serialQueue";
import type { MoveTarget } from "@/helpers/taskMoves";
import { toLocalDay } from "@/helpers/dates";
import type { Task } from "@/model";
import { useTagsStore, type TagOwnerWrite } from "@/stores/tags";
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

/** Orders sibling tasks by rank. */
export function compareByRank(a: Task, b: Task): number {
  return a.rank < b.rank ? -1 : a.rank > b.rank ? 1 : 0;
}

/**
 * Orders tasks as a depth-first tree, each parent followed by its subtasks,
 * siblings by rank. A task whose parent isn't in the list is shown at the
 * top level, so a filtered list never hides it.
 */
/**
 * Whether an open task's due date is before today, in the browser's time
 * zone.
 */
export function isTaskOverdue(task: Task, today: Date = new Date()): boolean {
  return !task.closed && !!task.dueDate && task.dueDate < toLocalDay(today);
}

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
    siblings.sort(compareByRank);
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

/**
 * What a task update does to cached task tags (see TagOwnerWrite), or
 * undefined when it leaves them alone. A task tag takes its name from its
 * task and its color from the task's regular tags, and is archived while
 * the task is closed. Closing or reopening a task cascades to its subtasks
 * or parents.
 */
function taskTagWrite(task: Task, patch: TaskPatch): TagOwnerWrite | undefined {
  const owner = { kind: "task", id: task.id } as const;
  if (patch.closed !== undefined) return { owner, cascades: true };
  if (patch.name !== undefined || patch.tagIds !== undefined) return { owner, tagId: task.tagId };
  return undefined;
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

    // Numbers each fresh fetch, so its key matches no other fetch's.
    let freshFetches = 0;

    /**
     * Fetches every task (closed ones too when includeClosed is set), with
     * its total time, replacing the local list. Joins a matching fetch
     * already in flight unless fresh is set, as it must be after a write: a
     * fetch sent before the write landed doesn't show it. Resolves once the
     * list shows this fetch or a newer one that superseded it.
     */
    async function fetchTasks({ fresh = false }: { fresh?: boolean } = {}): Promise<void> {
      const key = `all:${includeClosed.value}` + (fresh ? `:${++freshFetches}` : "");
      await supersededFetch.run(key, async () => {
        const result = await api.listAllTasks({ includeClosed: includeClosed.value });
        if (supersededFetch.isStale(key)) return;
        tasks.value = new Map(result.map((task) => [task.id, task]));
      });
      if (supersededFetch.isStale(key)) await supersededFetch.settled();
    }

    /** Reloads the list after a write (see fetchTasks). */
    function reloadAfterWrite(): Promise<void> {
      return fetchTasks({ fresh: true });
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
     * parents on the server, so the individually-fetched cache is dropped
     * (a fresh fetch is needed to see those cascading effects) and the list
     * is reloaded afterwards, unless reloadList is false because the caller
     * shows its own list and reloads that instead. Cached task tags the
     * update affects are refetched (see taskTagWrite).
     */
    async function updateTask(
      id: string,
      patch: TaskPatch,
      { reloadList = true }: { reloadList?: boolean } = {},
    ): Promise<Task> {
      const updated = await api.updateTask(id, patch);
      const tagWrite = taskTagWrite(updated, patch);
      if (tagWrite) void useTagsStore().ownerWritten(tagWrite);
      if (patch.closed !== undefined || patch.closeReason !== undefined) {
        individuallyFetchedTasks.value.delete(id);
        if (reloadList) await reloadAfterWrite();
        return copyTask(updated);
      }

      if (tasks.value.has(id)) {
        tasks.value.set(id, { ...updated, totalTimeMs: tasks.value.get(id)?.totalTimeMs });
      }
      const individual = individuallyFetchedTasks.value.get(id);
      if (individual) {
        individuallyFetchedTasks.value.set(id, {
          ...updated,
          totalTimeMs: individual.totalTimeMs,
          projectIds: individual.projectIds,
        });
      }
      return copyTask(updated);
    }

    async function closeTask(id: string, reason: Task["closeReason"] = "done"): Promise<Task> {
      return await updateTask(id, { closed: true, closeReason: reason });
    }

    async function reopenTask(id: string): Promise<Task> {
      return await updateTask(id, { closed: false });
    }

    // Serializes moves so a second one always sees the first one's fully
    // applied result (both its server move and the reload that follows)
    // before computing its own target - otherwise it can compute its target
    // from a stale order.
    const serializeTaskMove = createSerialQueue();

    /**
     * Moves a task to where target says, then reloads: the full list, or
     * reload when given (a list that loads its own tasks). target runs when
     * this move's turn comes, after the move before it has reloaded, so it
     * sees the order that move left; a null target skips the move. Resolves
     * to whether the task moved.
     */
    function moveTask(
      id: string,
      target: () => MoveTarget | null,
      { reload = reloadAfterWrite }: { reload?: () => Promise<void> } = {},
    ): Promise<boolean> {
      return serializeTaskMove(async () => {
        const to = target();
        if (!to) return false;
        await api.moveTask(id, to.parentId, to.afterTaskId);
        await reload();
        return true;
      });
    }

    /**
     * Deletes a task and its subtasks, along with their task tags. The
     * server refuses (409) when any of them has logged time.
     */
    async function deleteTask(id: string): Promise<void> {
      await api.deleteTask(id);
      void useTagsStore().ownerWritten({
        owner: { kind: "task", id },
        deleted: true,
        cascades: true,
      });
      individuallyFetchedTasks.value.delete(id);
      await reloadAfterWrite();
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
      deleteTask,
    };
  });
}

export const useTasksStore = createTasksStore(tasksApi);
export const __test__ = { createTasksStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTasksStore, import.meta.hot));
}
