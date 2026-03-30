<template>
  <div class="timesheet-row">
    <TagListEmbedded v-model="model.tagIds" />
    <div class="right-side">
      <TimespanEdit v-model="model" />
      <button class="icon-button" @click="toggleMenu" :aria-expanded="showMenu">
        <MaterialIcon class="centered-text more-icon" icon="more_horiz" size="1.5em" />
      </button>
    </div>
  </div>

  <div v-if="showMenu" class="timesheet-row menu-row">
    <div class="right-side">
      <button class="delete-button" @click="deleteThisTimespan">
        <MaterialIcon class="centered-text delete-icon" icon="delete" size="1.5em" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import TimespanEdit from "@/components/inputs/TimespanEdit.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import type { Timespan } from "@/model";
import { newTimespanWithDefaults } from "@/helpers/timespan";
import { ref } from "vue";
import { useTimespansStore } from "@/stores/timespans";

const timespansStore = useTimespansStore();

const model = defineModel<Timespan>({
  default: newTimespanWithDefaults(),
});

const showMenu = ref(false);

function toggleMenu() {
  showMenu.value = !showMenu.value;
}

function deleteThisTimespan() {
  timespansStore.deleteTimespan(model.value.id);
}
</script>

<style scoped>
.timesheet-row {
  display: flex;
}

.right-side {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.centered-text {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1em;
}

.delete-icon,
.more-icon {
  cursor: pointer;
  margin-left: 0.25em;
  margin-right: 0.25em;
}

.delete-icon {
  color: var(--nord11);
}

input[type="date"] {
  width: 11.5em;
  margin-left: 1em;
}

.context-menu {
  border: 1px solid red;
  position: absolute;
}

.delete-button {
  padding: 0.25em 0.5em;
}

.menu-row {
  margin: 0.25em;
}
</style>
