<template>
  <div class="timesheet-view">
    <div class="title-bar">
      <h2>Timesheet</h2>
    </div>
    <TheTimesheet />
  </div>
</template>

<script setup lang="ts">
import TheTimesheet from "@/components/timesheet/TheTimesheet.vue";
</script>

<style scoped>
.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: 1.25em;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
  padding: 0.25em 0;
}
</style>
