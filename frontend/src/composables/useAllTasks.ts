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
    reload: () => load(() => tasksStore.fetchTasks()),
    create: async (name) => {
      await tasksStore.createTaskFromName(name);
    },
  };
}
