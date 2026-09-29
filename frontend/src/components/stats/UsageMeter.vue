<template>
  <div class="meter">
    <div class="meter-header">
      <span class="meter-label">{{ label }}</span>
      <span class="meter-reading" :class="severityClass">
        <MaterialIcon
          v-if="severity === 'critical'"
          icon="warning"
          size="16px"
          class="meter-icon"
        />
        {{ format(used) }} / {{ format(limit) }} ({{ percentLabel }})
      </span>
    </div>
    <div class="meter-track">
      <div class="meter-fill" :class="severityClass" :style="{ width: fillPercent + '%' }" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";

/**
 * A labeled progress bar for a used/limit pair, e.g. a project's time budget
 * or a task's time estimate. The fill changes color as it approaches or
 * exceeds the limit. `used` and `limit` share a unit; `format` renders them.
 */
const props = defineProps<{
  label: string;
  used: number;
  limit: number;
  format: (value: number) => string;
}>();

// A limit of 0 (or less) is still a limit: it reads as full once anything is
// used, not as "no limit".
const ratio = computed(() => {
  if (props.limit <= 0) return props.used > 0 ? 1 : 0;
  return props.used / props.limit;
});

// "critical" once over the limit, "warning" from 80% up, "good" below that.
const severity = computed(() => {
  if (ratio.value > 1) return "critical";
  if (ratio.value >= 0.8) return "warning";
  return "good";
});
const severityClass = computed(() => `severity-${severity.value}`);
const fillPercent = computed(() => Math.min(ratio.value * 100, 100));
const percentLabel = computed(() => `${Math.round(ratio.value * 100)}%`);
</script>

<style scoped>
.meter {
  flex: 1;
  min-width: 16em;
  background-color: var(--nord1);
  border-radius: var(--radius-md);
  padding: 0.85em 1.25em;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.5em;
}

.meter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75em;
  font-size: 0.85em;
}

.meter-label {
  color: var(--nord4);
}

.meter-reading {
  display: inline-flex;
  align-items: center;
  gap: 0.25em;
  font-weight: 600;
  color: var(--nord4);
}

.meter-reading.severity-critical {
  color: var(--nord11);
}

.meter-icon {
  color: var(--nord11);
}

.meter-track {
  position: relative;
  height: 10px;
  border-radius: 999px;
  background-color: var(--nord3);
  overflow: hidden;
}

.meter-fill {
  height: 100%;
  border-radius: 999px;
  transition: width var(--transition-base, 0.2s ease);
}

.meter-fill.severity-good {
  background-color: var(--nord14);
}

.meter-fill.severity-warning {
  background-color: var(--nord13);
}

.meter-fill.severity-critical {
  background-color: var(--nord11);
}
</style>
