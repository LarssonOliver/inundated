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
      @keyup.control="isCtrlDown = false"
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
      @keyup.control="isCtrlDown = false"
      @focusout="close"
    >
      {{ selectedLabel || placeholder }}
    </button>

    <MaterialIcon
      v-if="!creatable"
      icon="expand_more"
      size="1.2em"
      class="dropdown-chevron"
      :class="{ open: isOpen }"
    />

    <ul v-if="isOpen && hasPanelContent" class="dropdown-panel" role="listbox">
      <li
        v-for="(option, index) in filteredOptions"
        :key="optionValue(option)"
        role="option"
        :aria-selected="optionValue(option) === modelValue"
        :class="{ highlighted: index === highlightedIndex }"
        @mousedown.prevent="selectOption(option)"
        @mouseenter="highlightedIndex = index"
      >
        <slot :option="option">{{ optionLabel(option) }}</slot>
      </li>
      <template v-if="creatable && filterText">
        <hr v-if="filteredOptions.length > 0" />
        <li
          data-testid="create-row"
          :class="{ highlighted: filteredOptions.length === highlightedIndex }"
          @mousedown.prevent="createItem"
          @mouseenter="highlightedIndex = filteredOptions.length"
        >
          <slot name="create" :query="filterText">Create "{{ filterText }}"...</slot>
        </li>
      </template>
    </ul>
  </div>
</template>

<script setup lang="ts" generic="T">
import { computed, ref } from "vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";

export interface DropdownOption {
  value: string;
  label: string;
}

const props = withDefaults(
  defineProps<{
    options: T[];
    searchable?: boolean;
    placeholder?: string;
    manualFilter?: boolean;
    creatable?: boolean;
    optionValue?: (option: T) => string;
    optionLabel?: (option: T) => string;
  }>(),
  {
    optionValue: (option: T) => (option as DropdownOption).value,
    optionLabel: (option: T) => (option as DropdownOption).label,
  },
);

const emit = defineEmits<{
  search: [query: string];
  create: [query: string];
  select: [option: T];
}>();

const modelValue = defineModel<string>({ default: "" });

const isOpen = ref(false);
const highlightedIndex = ref<number | null>(null);
const filterText = ref("");
const isCtrlDown = ref(false);

const selectedOption = computed(() =>
  props.options.find((option) => props.optionValue(option) === modelValue.value),
);
const selectedLabel = computed(() =>
  selectedOption.value ? props.optionLabel(selectedOption.value) : "",
);

const filteredOptions = computed(() => {
  if (props.manualFilter || !props.searchable || !filterText.value) {
    return props.options;
  }
  const query = filterText.value.toLowerCase();
  return props.options.filter((option) => props.optionLabel(option).toLowerCase().includes(query));
});

const showCreateRow = computed(() => props.creatable && filterText.value.length > 0);

const hasPanelContent = computed(() => filteredOptions.value.length > 0 || showCreateRow.value);

function open() {
  if (isOpen.value) {
    return;
  }
  isOpen.value = true;
  filterText.value = "";
  const current = props.options.findIndex(
    (option) => props.optionValue(option) === modelValue.value,
  );
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
  setFilterText("");
}

function setFilterText(value: string) {
  filterText.value = value;
  if (props.manualFilter) {
    emit("search", value);
  }
}

function onFilterInput(event: Event) {
  isOpen.value = true;
  setFilterText((event.target as HTMLInputElement).value);
  highlightedIndex.value = null;
}

function selectOption(option: T) {
  modelValue.value = props.optionValue(option);
  emit("select", option);
  close();
}

function createItem() {
  if (!showCreateRow.value) {
    return;
  }
  emit("create", filterText.value);
  close();
}

function moveHighlight(delta: 1 | -1) {
  const maxIndex = filteredOptions.value.length - (showCreateRow.value ? 0 : 1);
  if (maxIndex < 0) {
    return;
  }
  if (highlightedIndex.value === null) {
    if (delta > 0) {
      highlightedIndex.value = 0;
    }
    return;
  }
  const next = highlightedIndex.value + delta;
  if (next < 0) {
    highlightedIndex.value = props.creatable ? null : 0;
    return;
  }
  highlightedIndex.value = Math.min(next, maxIndex);
}

function selectHighlighted() {
  if (highlightedIndex.value !== null && highlightedIndex.value < filteredOptions.value.length) {
    selectOption(filteredOptions.value[highlightedIndex.value]);
  } else if (
    props.creatable &&
    (highlightedIndex.value === null || highlightedIndex.value === filteredOptions.value.length)
  ) {
    createItem();
  }
}

function handleKeydown(event: KeyboardEvent) {
  switch (event.key) {
    case "Control":
      isCtrlDown.value = true;
      break;
    case "n":
      if (!isCtrlDown.value) break;
    case "ArrowDown":
      if (!isOpen.value) {
        open();
      } else {
        moveHighlight(1);
      }
      event.preventDefault();
      break;
    case "p":
      if (!isCtrlDown.value) break;
    case "ArrowUp":
      if (isOpen.value) {
        moveHighlight(-1);
      }
      event.preventDefault();
      break;
    case "u":
      if (!isCtrlDown.value) break;
      setFilterText("");
      highlightedIndex.value = null;
      break;
    case "y":
      if (!isCtrlDown.value) break;
    case "Enter":
      if (isOpen.value) {
        selectHighlighted();
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
  max-width: var(--max-width, 400px);
}

.dropdown-trigger {
  text-align: left;
  cursor: pointer;
  padding-right: 2em;
}

.dropdown-chevron {
  position: absolute;
  top: 50%;
  right: 0.6em;
  transform: translateY(-50%);
  color: var(--nord3);
  pointer-events: none;
  transition: transform var(--transition-fast);
}

.dropdown-chevron.open {
  transform: translateY(-50%) rotate(180deg);
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
  border: 1px solid var(--nord1);
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

.dropdown-panel hr {
  border: none;
  border-top: 1px solid var(--nord1);
  margin: 0.35em 0;
}
</style>
