<template>
  <div class="timesheet-row">
    <TagListEmbedded v-model="tagIds" />
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
import { computed, ref } from "vue";
import { useTimespansStore } from "@/stores/timespans";

const timespansStore = useTimespansStore();

const model = defineModel<Timespan>({
  default: newTimespanWithDefaults(),
});

const tagIds = computed({
  get: () => model.value.tagIds,
  set: (value: Set<string>) => {
    model.value = { ...model.value, tagIds: value };
  },
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
  padding: 0.4em 0.5em;
  border-radius: var(--radius-sm);
  transition: background-color var(--transition-fast);
}

.timesheet-row:hover {
  background-color: var(--nord1);
}

.right-side {
  display: flex;
  align-items: center;
  margin-left: auto;
  margin-right: 0.5em;
}

.centered-text {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1em;
}

.icon-button {
  width: auto;
  max-width: none;
  padding: 0.35em;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 0.25em;
}

.icon-button:hover {
  background-color: var(--nord2);
}

.delete-icon {
  color: var(--nord11);
}

input[type="date"] {
  width: 11.5em;
  margin-left: 1em;
}

.delete-button {
  width: auto;
  max-width: none;
  padding: 0.35em 0.5em;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
}

.delete-button:hover {
  background-color: var(--nord2);
}

.menu-row {
  margin: 0.25em 0.5em;
}
</style>
