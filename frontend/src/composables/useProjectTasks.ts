import { tasksApi } from "@/api";
import type { Project, Task } from "@/model";
import { useTasksStore } from "@/stores/tasks";
import { computed, ref, watch } from "vue";
import type { TaskListSource } from "./taskListSource";

/**
 * A project's tasks (the project page): those carrying one of its tags,
 * those whose own tag it links, and their subtasks. project is the project
 * as last saved, since that's what the server lists tasks by. remove takes
 * a task's own tag out of the project, which the page owns.
 */
export function useProjectTasks(
  project: () => Project,
  { remove }: { remove: (task: Task) => Promise<void> },
): TaskListSource {
  const tasksStore = useTasksStore();
  const tasks = ref<Task[]>([]);
  const showClosed = ref(false);
  const isLoading = ref(false);
  const loadFailed = ref(false);

  // Loads can overlap (the project is replaced several times while the page
  // opens, and after every save); only the latest one may land.
  let loadSeq = 0;

  async function reload() {
    const { id } = project();
    if (!id) return;
    const seq = ++loadSeq;
    isLoading.value = true;
    try {
      const result = await tasksApi.listAllTasks({
        projectId: id,
        includeClosed: showClosed.value,
      });
      if (seq !== loadSeq) return;
      tasks.value = result;
      loadFailed.value = false;
    } catch {
      if (seq === loadSeq) loadFailed.value = true;
    } finally {
      if (seq === loadSeq) isLoading.value = false;
    }
  }

  // Changing the project's tags changes which tasks belong to it. The key is
  // a string so that replacing the project with an equal copy doesn't
  // reload.
  watch(
    () => `${project().id}:${[...project().tagIds].sort().join(",")}:${showClosed.value}`,
    () => void reload(),
    { immediate: true },
  );

  // Looked up for every listed task when TaskList works out its move buttons.
  const listedIds = computed(() => new Set(tasks.value.map((task) => task.id)));

  /** Whether the task is in the project through its own tags, not just an ancestor's. */
  function inProjectOnItsOwn(task: Task): boolean {
    const { tagId, tagIds } = project();
    return tagIds.has(task.tagId) || [...task.tagIds].some((id) => id === tagId || tagIds.has(id));
  }

  return {
    tasks,
    showClosed,
    setShowClosed: (value) => {
      showClosed.value = value;
    },
    isLoading,
    loadFailed,
    reload,
    // A new task joins the project by carrying its project tag, so there's
    // no way to create one (and TaskList shows no add box) without it.
    get create() {
      const projectTagId = project().tagId;
      if (!projectTagId) return undefined;
      return async (name: string) => {
        await tasksStore.createTask({ name, tagIds: new Set([projectTagId]) });
        await reload();
      };
    },
    // A task in the project only through its parent leaves it when moved
    // out from under a listed task.
    keepsInList: (task, newParentId) =>
      inProjectOnItsOwn(task) || (newParentId !== undefined && listedIds.value.has(newParentId)),
    removeLabel: "Remove from project",
    canRemove: (task) => project().tagIds.has(task.tagId),
    remove,
  };
}
