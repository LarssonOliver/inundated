<template>
  <div class="timesheet-row">
    <TagListEmbedded
      v-model="model.tagIds"
      @update:model-value="() => $emit('update:model-value', model)"
    />
    <div class="right-side">
      <TimeSpanEdit
        :model-value="model"
        @update:model-value="(value) => $emit('update:model-value', value)"
      />
      <MaterialIcon
        class="centered-text more-icon"
        icon="more_horiz"
        size="1.5em"
        @mousedown="contextMenuClick"
      />
    </div>
  </div>
  <div v-if="showMenu" class="timesheet-row menu-row">
    <div class="right-side">
      <button class="delete-button">
        <MaterialIcon
          class="centered-text delete-icon"
          icon="delete"
          size="1.5em"
          @mousedown="deleteTimeSpan"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import TimeSpanEdit from "@/components/inputs/TimeSpanEdit.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import type { TimeSpan } from "@/model/model";
import { newTimespanWithDefaults } from "@/helpers/timespan";
import { ref } from "vue";
import { useTimeSpansStore } from "@/stores/timespans";

const timeSpansStore = useTimeSpansStore();

const model = defineModel<TimeSpan>({
  default: newTimespanWithDefaults(),
});

const showMenu = ref(false);

defineEmits<{
  "update:model-value": [value: TimeSpan];
}>();

function contextMenuClick() {
  showMenu.value = !showMenu.value;
}

function deleteTimeSpan() {
  const index = timeSpansStore.timeSpans.findIndex((ts) => ts.id === model.value.id);
  if (index < 0) return;
  timeSpansStore.timeSpans.splice(index, 1);
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
