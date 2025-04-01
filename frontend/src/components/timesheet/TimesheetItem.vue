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
        @mousedown="console.log('TODO')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import TimeSpanEdit from "@/components/inputs/TimeSpanEdit.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import type { TimeSpan } from "@/model/model";
import { newTimespanWithDefaults } from "@/helpers/timespan";

const model = defineModel<TimeSpan>({
  default: newTimespanWithDefaults(),
});

defineEmits<{
  "update:model-value": [value: TimeSpan];
}>();
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

.more-icon {
  cursor: pointer;
  margin-left: 0.25em;
  margin-right: 0.25em;
}

input[type="date"] {
  width: 11.5em;
  margin-left: 1em;
}
</style>
