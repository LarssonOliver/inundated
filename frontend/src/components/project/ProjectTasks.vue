<template>
  <TaskList
    class="project-tasks"
    :source="source"
    add-placeholder="Add a task to this project..."
    empty-text="No tasks yet. Tasks join this project when they carry one of its tags, or when their own tag is added to the project."
  >
    <template #title>
      <div class="header">
        <h2>Tasks</h2>
        <div v-if="project.totalTimeMs !== undefined" class="time-split">
          <span title="Time logged on this project's tasks">
            On tasks <strong>{{ formatMs(taskTimeMs) }}</strong>
          </span>
          <span title="Project time not logged on any of its tasks">
            Additional <strong>{{ formatMs(additionalTimeMs) }}</strong>
          </span>
        </div>
      </div>
    </template>
  </TaskList>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Project, Task } from "@/model";
import { useSettingsStore } from "@/stores/settings";
import { useDurationFormat } from "@/composables/useDurationFormat";
import { useProjectTasks } from "@/composables/useProjectTasks";
import TaskList from "@/components/tasks/TaskList.vue";

const props = defineProps<{
  /** The project as last saved, which the server lists its tasks by. */
  project: Project;
}>();

/** The page removes a task's own tag from the project (see ProjectView). */
const emit = defineEmits<{ "remove-task": [task: Task] }>();

const settingsStore = useSettingsStore();
const formatMs = useDurationFormat(() => settingsStore.settings);

const source = useProjectTasks(() => props.project, {
  remove: async (task) => emit("remove-task", task),
});

const taskTimeMs = computed(() => props.project.taskTimeMs ?? 0);
const additionalTimeMs = computed(() =>
  Math.max(0, (props.project.totalTimeMs ?? 0) - taskTimeMs.value),
);
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1em;
}

.time-split {
  display: flex;
  gap: 1.5em;
  color: var(--nord3);
}

.time-split strong {
  color: var(--nord4);
}

.project-tasks :deep(.task-row) {
  background-color: var(--nord1);
  box-shadow: none;
}
</style>
