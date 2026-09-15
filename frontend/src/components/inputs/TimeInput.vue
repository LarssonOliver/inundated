<template>
  <div>
    <input
      type="text"
      v-model="currentValue"
      placeholder="00:00"
      :class="{ 'time-12h': format === '12h' }"
      @keydown.enter="valueEntered"
      @focusout="valueEntered"
      @focus="($event.target as HTMLInputElement).select()"
      @onmouseup.prevent
    />
    <sup v-if="showNextDay">+1</sup>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { formatClockTime, parseClockTime, parseGoDuration } from "@/helpers/time";
import type { TimeFormat } from "@/model";

const props = defineProps<{
  showNextDay?: boolean;
  resolveDuration?: (durationMs: number) => string | null;
  timeFormat?: TimeFormat;
}>();

// The exposed model is always canonical 24h "HH:MM" - callers (and
// resolveDuration) depend on that shape. Only what's displayed/typed in the
// input itself is rendered per the timeFormat prop.
const model = defineModel<string>({ default: "00:00" });
const format = computed(() => props.timeFormat ?? "24h");

function toDisplay(canonical: string): string {
  const [hours, minutes] = canonical.split(":").map((s) => +s);
  return formatClockTime(hours, minutes, format.value);
}

function toCanonical(hours: number, minutes: number): string {
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

// model is always kept valid (it only ever holds a well-formed canonical
// time), so it doubles as its own "last valid value" - no separate ref
// needed to remember what to revert to.
const currentValue = ref<string>(toDisplay(model.value));

watch(model, (newValue) => {
  currentValue.value = toDisplay(newValue);
});

watch(format, () => {
  currentValue.value = toDisplay(model.value);
});

function resetToLastValid() {
  currentValue.value = toDisplay(model.value);
}

function valueEntered() {
  const parsed = parseClockTime(currentValue.value);

  if (parsed === null) {
    const durationMs = parseGoDuration(currentValue.value);
    const resolved = durationMs === null ? null : (props.resolveDuration?.(durationMs) ?? null);
    if (resolved === null) {
      resetToLastValid();
      return;
    }
    model.value = resolved;
    currentValue.value = toDisplay(resolved);
    return;
  }

  const canonical = toCanonical(parsed.hours, parsed.minutes);
  model.value = canonical;
  currentValue.value = toDisplay(canonical);
}
</script>

<style scoped>
input {
  width: 4.8em;
  font-family: monospace;
}

input.time-12h {
  width: 6.5em;
}

sup {
  font-size: 0.6em;
  position: relative;
  margin-left: -2em;
  font-family: monospace;
}
</style>
