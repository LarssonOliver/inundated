<template>
  <div class="timesheet-row">
    <TagListEmbedded
      v-model="model.tagIds"
      @update:model-value="() => $emit('update:model-value', model)"
    />
    <div class="right-side">
      <TimespanEdit
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
          @mousedown="deleteThisTimespan"
        />
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

defineEmits<{
  "update:model-value": [value: Timespan];
}>();

function contextMenuClick() {
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
