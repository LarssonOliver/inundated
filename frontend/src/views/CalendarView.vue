<template>
  <div class="calendar-view">
    <div class="title-bar">
      <h2>Calendar</h2>
    </div>
    <TheCalendar class="calendar-body" />
  </div>
</template>

<script setup lang="ts">
import TheCalendar from "@/components/calendar/TheCalendar.vue";
</script>

<style scoped>
.calendar-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: 1.25em;
  flex-shrink: 0;
}

.calendar-body {
  flex: 1;
  min-height: 0;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
  padding: 0.25em 0;
}
</style>
