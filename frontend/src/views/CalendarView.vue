<template>
  <div class="calendar-view">
    <div class="title-bar">
      <h2>Calendar</h2>
    </div>
    <TheCalendar />
  </div>
</template>

<script setup lang="ts">
import TheCalendar from "@/components/calendar/TheCalendar.vue";
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
