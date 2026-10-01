import { tasksApi } from "@/api";
import type { Task } from "@/model";
import { useTasksStore } from "@/stores/tasks";
import { computed, ref, watch } from "vue";
import type { TaskListSource } from "./taskListSource";

/** A task's subtasks, nested ones included (the task page). */
export function useSubtasks(taskId: () => string | undefined): TaskListSource {
  const tasksStore = useTasksStore();
  // Every task, closed ones included even while they aren't listed, so an
  // open task under a closed one still connects up to the task.
  const allTasks = ref<Task[]>([]);
  const showClosed = ref(false);
  const isLoading = ref(false);
  const loadFailed = ref(false);

  const descendants = computed(() => {
    const rootId = taskId();
    if (!rootId) return [];
    const byId = new Map(allTasks.value.map((task) => [task.id, task]));
    return allTasks.value.filter((task) => isDescendant(task, rootId, byId));
  });

  // Loads can overlap when the page moves to another task; only the latest
  // one may land.
  let loadSeq = 0;

  async function reload() {
    if (!taskId()) return;
    const seq = ++loadSeq;
    isLoading.value = true;
    try {
      const result = await tasksApi.listAllTasks({ includeClosed: true });
      if (seq !== loadSeq) return;
      allTasks.value = result;
      loadFailed.value = false;
    } catch {
      if (seq === loadSeq) loadFailed.value = true;
    } finally {
      if (seq === loadSeq) isLoading.value = false;
    }
  }

  watch(taskId, () => void reload(), { immediate: true });

  return {
    tasks: computed(() => descendants.value.filter((task) => showClosed.value || !task.closed)),
    showClosed,
    setShowClosed: (value) => {
      showClosed.value = value;
    },
    isLoading,
    loadFailed,
    reload,
    create: async (name) => {
      const parentId = taskId();
      if (parentId) await tasksStore.createTaskFromName(name, parentId);
    },
    // Outdenting a direct subtask would move it out of the tree.
    keepsInList: (_task, newParentId) =>
      newParentId !== undefined &&
      (newParentId === taskId() || descendants.value.some((task) => task.id === newParentId)),
  };
}

function isDescendant(task: Task, rootId: string, byId: Map<string, Task>): boolean {
  const seen = new Set<string>();
  let parentId = task.parentId;
  while (parentId && !seen.has(parentId)) {
    if (parentId === rootId) return true;
    seen.add(parentId);
    parentId = byId.get(parentId)?.parentId;
  }
  return false;
}
