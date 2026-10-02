import { useTasksStore } from "@/stores/tasks";
import { computed, ref } from "vue";
import type { TaskListSource } from "./taskListSource";

/** Every task, as the tasks store lists them (the Tasks page). */
export function useAllTasks(): TaskListSource {
  const tasksStore = useTasksStore();
  const loadFailed = ref(false);

  async function load(fetch: () => Promise<void>) {
    try {
      await fetch();
      loadFailed.value = false;
    } catch {
      loadFailed.value = true;
    }
  }

  void load(() => tasksStore.fetchTasks());

  return {
    tasks: computed(() => tasksStore.tasks),
    showClosed: computed(() => tasksStore.includeClosed),
    setShowClosed: (value) => void load(() => tasksStore.setIncludeClosed(value)),
    isLoading: computed(() => tasksStore.isLoading),
    loadFailed,
    // Called after a change, which a fetch already in flight wouldn't show.
    reload: () => load(() => tasksStore.fetchTasks({ fresh: true })),
    // The store updates its list in place, except that closing or reopening
    // a task also closes or reopens others on the server.
    reloadsAfter: (patch) => patch.closed !== undefined || patch.closeReason !== undefined,
    // The store lists the new task itself.
    create: async (name) => {
      await tasksStore.createTaskFromName(name);
    },
  };
}
