import { tasksApi, type NewTask, type TasksApi } from "@/api";
import type { TaskPatch } from "@/api/mappers";
import { useSupersededFetch } from "@/composables/useSupersededFetch";
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

    /**
     * Fetches a task's direct subtasks, closed ones included, bypassing the
     * list's includeClosed filter.
     */
    async function fetchSubtasks(parentId: string): Promise<Task[]> {
      const subtasks = await api.listAllTasks({ parentId, includeClosed: true });
      return subtasks.map(copyTask).sort(compareByRank);
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
        if (reloadList) await fetchTasks();
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

    async function moveTaskRaw(id: string, parentId?: string, afterTaskId?: string): Promise<Task> {
      const moved = await api.moveTask(id, parentId, afterTaskId);
      await fetchTasks();
      return copyTask(moved);
    }

    async function shiftTaskRaw(id: string, delta: -1 | 1): Promise<void> {
      const task = tasks.value.get(id);
      if (!task) return;
      // Open and closed tasks render as separate lists (see TaskListView), so
      // moving one must only ever reorder it among same-status siblings.
      const siblings = taskTree(readOnlyTasks.value)
        .map((row) => row.task)
        .filter((t) => t.parentId === task.parentId && t.closed === task.closed);
      const index = siblings.findIndex((t) => t.id === id);
      const target = index + delta;
      if (index === -1 || target < 0 || target >= siblings.length) return;

      // Moving up places it after the sibling two above (or first); moving
      // down places it after the next sibling.
      const others = siblings.filter((t) => t.id !== id);
      const after = target === 0 ? undefined : others[target - 1];
      await moveTaskRaw(id, task.parentId, after?.id);
    }

    /**
     * Makes a task the last child of the task immediately above it (within
     * its own open/closed section, depth-first order) - a standard outliner
     * "indent". A no-op if there's no row above it.
     */
    async function indentTaskRaw(id: string): Promise<void> {
      const task = tasks.value.get(id);
      if (!task) return;
      const sectionRows = taskTree(readOnlyTasks.value.filter((t) => t.closed === task.closed));
      const index = sectionRows.findIndex((row) => row.task.id === id);
      if (index <= 0) return;

      const newParent = sectionRows[index - 1].task;
      const lastChild = sectionRows
        .map((row) => row.task)
        .filter((t) => t.parentId === newParent.id && t.id !== id)
        .at(-1);
      await moveTaskRaw(id, newParent.id, lastChild?.id);
    }

    /**
     * Makes a task a sibling of its current parent, placed directly after
     * it - a standard outliner "outdent". A no-op on a top-level task.
     */
    async function outdentTaskRaw(id: string): Promise<void> {
      const task = tasks.value.get(id);
      if (!task?.parentId) return;
      const parent = tasks.value.get(task.parentId);
      // Closing only cascades parent -> child, so a closed task can have an
      // open parent; outdenting must still only ever place it among
      // same-status siblings (see shiftTaskRaw/indentTaskRaw above).
      if (!parent || parent.closed !== task.closed) return;
      await moveTaskRaw(id, parent.parentId, parent.id);
    }

    // Serializes moveTask/shiftTask so a second call always sees the first
    // one's fully-applied result (both its server move and the refetch that
    // follows) before computing its own target or re-fetching - otherwise a
    // second move can compute its target from a stale local order, and its
    // own fetchTasks() call can be deduped away (as a duplicate of the
    // still-in-flight first one) by useSupersededFetch, silently dropping
    // its result.
    let taskMoveQueue: Promise<void> = Promise.resolve();

    function serializeTaskMove<T>(fn: () => Promise<T>): Promise<T> {
      const run = taskMoveQueue.then(fn, fn);
      taskMoveQueue = run.then(
        () => undefined,
        () => undefined,
      );
      return run;
    }

    /**
     * Moves a task under parentId (top level when unset), directly after
     * afterTaskId (first when unset). Siblings may be re-ranked, so the list
     * is reloaded afterwards.
     */
    function moveTask(id: string, parentId?: string, afterTaskId?: string): Promise<Task> {
      return serializeTaskMove(() => moveTaskRaw(id, parentId, afterTaskId));
    }

    /**
     * Moves a task one step up or down among its siblings in the list.
     */
    function shiftTask(id: string, delta: -1 | 1): Promise<void> {
      return serializeTaskMove(() => shiftTaskRaw(id, delta));
    }

    /** See indentTaskRaw. */
    function indentTask(id: string): Promise<void> {
      return serializeTaskMove(() => indentTaskRaw(id));
    }

    /** See outdentTaskRaw. */
    function outdentTask(id: string): Promise<void> {
      return serializeTaskMove(() => outdentTaskRaw(id));
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
      fetchSubtasks,
      createTask,
      createTaskFromName,
      updateTask,
      closeTask,
      reopenTask,
      moveTask,
      shiftTask,
      indentTask,
      outdentTask,
      deleteTask,
    };
  });
}

export const useTasksStore = createTasksStore(tasksApi);
export const __test__ = { createTasksStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTasksStore, import.meta.hot));
}
