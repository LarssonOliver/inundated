<template>
  <div class="container">
    <TimesheetItem />
  </div>
</template>

<script setup lang="ts">
import TimesheetItem from "@/components/timesheet/TimesheetItem.vue";
</script>

<style scoped>
.container {
  width: 100%;
  border: 1px solid var(--nord3);
  min-height: 100px;
}
</style>
