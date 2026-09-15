<template>
  <div class="dropdown">
    <input
      v-if="searchable"
      type="text"
      class="dropdown-trigger"
      role="combobox"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
      :placeholder="placeholder"
      :value="isOpen ? filterText : selectedLabel"
      @focus="open"
      @input="onFilterInput"
      @keydown="handleKeydown"
      @focusout="close"
    />
    <button
      v-else
      type="button"
      class="dropdown-trigger"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
      @click="toggleOpen"
      @keydown="handleKeydown"
      @focusout="close"
    >
      {{ selectedLabel || placeholder }}
    </button>

    <ul v-if="isOpen" class="dropdown-panel" role="listbox">
      <li
        v-for="(option, index) in filteredOptions"
        :key="option.value"
        role="option"
        :aria-selected="option.value === modelValue"
        :class="{ highlighted: index === highlightedIndex }"
        @mousedown.prevent="selectOption(option)"
        @mouseenter="highlightedIndex = index"
      >
        {{ option.label }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

export interface DropdownOption {
  value: string;
  label: string;
}

const props = defineProps<{
  options: DropdownOption[];
  searchable?: boolean;
  placeholder?: string;
}>();

const modelValue = defineModel<string>({ default: "" });

const isOpen = ref(false);
const highlightedIndex = ref<number | null>(null);
const filterText = ref("");

const selectedOption = computed(() =>
  props.options.find((option) => option.value === modelValue.value),
);
const selectedLabel = computed(() => selectedOption.value?.label ?? "");

const filteredOptions = computed(() => {
  if (!props.searchable || !filterText.value) {
    return props.options;
  }
  const query = filterText.value.toLowerCase();
  return props.options.filter((option) => option.label.toLowerCase().includes(query));
});

function open() {
  if (isOpen.value) {
    return;
  }
  isOpen.value = true;
  filterText.value = "";
  const current = props.options.findIndex((option) => option.value === modelValue.value);
  highlightedIndex.value = current === -1 ? null : current;
}

function toggleOpen() {
  if (isOpen.value) {
    close();
  } else {
    open();
  }
}

function close() {
  isOpen.value = false;
  highlightedIndex.value = null;
  filterText.value = "";
}

function onFilterInput(event: Event) {
  filterText.value = (event.target as HTMLInputElement).value;
  highlightedIndex.value = null;
}

function selectOption(option: DropdownOption) {
  modelValue.value = option.value;
  close();
}

function handleKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case "ArrowDown":
      if (!isOpen.value) {
        open();
      } else {
        highlightedIndex.value = Math.min(
          (highlightedIndex.value ?? -1) + 1,
          filteredOptions.value.length - 1,
        );
      }
      event.preventDefault();
      break;
    case "ArrowUp":
      if (isOpen.value && highlightedIndex.value !== null) {
        highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0);
      }
      event.preventDefault();
      break;
    case "Enter":
      if (isOpen.value && highlightedIndex.value !== null) {
        selectOption(filteredOptions.value[highlightedIndex.value]);
        event.preventDefault();
      }
      break;
    case "Escape":
      close();
      (event.target as HTMLElement).blur();
      break;
    default:
      break;
  }
}
</script>

<style scoped>
.dropdown {
  position: relative;
  width: 100%;
  max-width: 400px;
}

.dropdown-trigger {
  text-align: left;
  cursor: pointer;
}

.dropdown-panel {
  position: absolute;
  top: calc(100% + 0.25em);
  left: 0;
  right: 0;
  max-height: 16em;
  overflow-y: auto;
  background-color: var(--nord0);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: 0.35em;
  margin: 0;
  list-style-type: none;
  z-index: 1;
}

.dropdown-panel li {
  padding: 0.4em 0.5em;
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.dropdown-panel li.highlighted {
  background-color: var(--nord3);
}
</style>
