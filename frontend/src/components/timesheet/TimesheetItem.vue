<template>
  <div class="timesheet-row">
    Test
    <!-- <input type="date" /> -->
    <!-- <input class="add-button" type="button" value="Button" /> -->
    <TimeInput />
  </div>
</template>

<script setup lang="ts">
import TimeInput from "@/components/inputs/TimeInput.vue";
</script>

<style scoped>
.timesheet-row {
  border: 1px solid var(--nord3);
  display: flex;
}

/* .add-button { */
/*   background-color: var(--nord8); */
/*   color: var(--nord0); */
/*   width: auto; */
/*   margin-left: auto; */
/* } */

/* .add-button:hover { */
/*   background-color: rgb(121, 184, 202); */
/* } */
</style>
