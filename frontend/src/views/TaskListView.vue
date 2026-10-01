<template>
  <TaskList
    :source="source"
    empty-text='No tasks yet. Add one above, or type "#" and a name in any tag field.'
    empty-open-text='No open tasks. Add one above, or type "#" and a name in any tag field.'
  >
    <template #title><h2>Tasks</h2></template>
  </TaskList>
</template>

<script setup lang="ts">
import TaskList from "@/components/tasks/TaskList.vue";
import { useAllTasks } from "@/composables/useAllTasks";

const source = useAllTasks();
</script>

<style scoped>
h2 {
  margin: 0;
}
</style>
