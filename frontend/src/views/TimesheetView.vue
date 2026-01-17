<template>
  <div class="timesheet-view">
    <h2>Timesheet</h2>
    <TheTimesheet />
  </div>
</template>

<script setup lang="ts">
import TheTimesheet from "@/components/timesheet/TheTimesheet.vue";
</script>

<style scoped>
.timesheet-view {
  margin: 0 1em;
}
</style>
