import { tasksApi } from "@/api";
import type { Task } from "@/model";
import { useTasksStore } from "@/stores/tasks";
import { computed, ref, watch } from "vue";
import type { TaskListSource } from "./taskListSource";

/** A task's subtasks, nested ones included (the task page). */
export function useSubtasks(taskId: () => string | undefined): TaskListSource {
  const tasksStore = useTasksStore();
  // The task's whole subtree, closed tasks included even while they aren't
  // listed, so an open task under a closed one still connects up to the task.
  const subtree = ref<Task[]>([]);
  const showClosed = ref(false);
  const isLoading = ref(false);
  const loadFailed = ref(false);

  // Loads can overlap when the page moves to another task; only the latest
  // one may land.
  let loadSeq = 0;

  async function reload() {
    const rootId = taskId();
    if (!rootId) return;
    const seq = ++loadSeq;
    isLoading.value = true;
    try {
      const result = await tasksApi.listAllTasks({ ancestorId: rootId, includeClosed: true });
      if (seq !== loadSeq) return;
      subtree.value = result;
      loadFailed.value = false;
    } catch {
      if (seq === loadSeq) loadFailed.value = true;
    } finally {
      if (seq === loadSeq) isLoading.value = false;
    }
  }

  // Looked up for every listed task when TaskList works out its move buttons.
  const subtreeIds = computed(() => new Set(subtree.value.map((task) => task.id)));

  // Each task's subtasks start out as a freshly opened page shows them.
  watch(
    taskId,
    (_id, previousId) => {
      if (previousId !== undefined) {
        showClosed.value = false;
        // The previous task's subtree must not show under this one.
        subtree.value = [];
      }
      void reload();
    },
    { immediate: true },
  );

  return {
    tasks: computed(() => subtree.value.filter((task) => showClosed.value || !task.closed)),
    showClosed,
    setShowClosed: (value) => {
      showClosed.value = value;
    },
    isLoading,
    loadFailed,
    reload,
    create: async (name) => {
      const parentId = taskId();
      if (!parentId) return;
      await tasksStore.createTaskFromName(name, parentId);
      await reload();
    },
    // Outdenting a direct subtask would move it out of the tree.
    keepsInList: (_task, newParentId) =>
      newParentId !== undefined && (newParentId === taskId() || subtreeIds.value.has(newParentId)),
  };
}
