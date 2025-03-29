<template>
  <div class="search-container">
    <input
      type="text"
      :placeholder="props.placeholder"
      @mouseenter="showDropdown = true"
      @mouseleave="showDropdown = false"
      @focusin="showDropdown = true"
      @focusout="showDropdown = false"
      @keydown="handleKeyDown"
      @keyup.control="isCtrlDown = false"
      @input="onInput"
      v-model="searchString"
    />
    <div
      v-if="showDropdown"
      class="search-dropdown"
      @mouseenter="showDropdown = true"
      @mouseleave="showDropdown = false"
    >
      <div v-if="searching" class="searching-text">Searching...</div>
      <ul v-else>
        <li
          v-for="(item, index) of props.items || []"
          v-bind:key="index"
          @mouseenter="highlightedIndex = index"
          @mouseleave="highlightedIndex = null"
          @mousedown="selectItem(item)"
          :class="{ 'highlight-item': highlightedIndex === index }"
        >
          <slot v-bind="item" />
        </li>
        <hr v-if="showCreateItemField" />
        <div
          v-if="showCreateItemField"
          @mouseenter="highlightedIndex = props.items?.length || 0"
          @mouseleave="highlightedIndex = null"
          @mousedown="createItem()"
          :class="{ 'highlight-item': highlightedIndex === props.items?.length }"
        >
          <slot name="createItemField" :v-bind="searchString">
            <li>
              <span>Create "{{ searchString }}"...</span>
            </li>
          </slot>
        </div>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import { computed, ref, watch } from "vue";

const props = defineProps<{
  items?: T[];
  placeholder?: string;
}>();

const emit = defineEmits<{
  search: [query: string];
  create: [createName: string];
  select: [item: T];
}>();

const searchString = ref("");
const highlightedIndex = ref<number | null>(null);

const showDropdown = ref(false);
const searching = ref(false);
const isCtrlDown = ref(false);

const showCreateItemField = computed(() => {
  return searchString.value.length > 0;
});

function onInput() {
  highlightedIndex.value = null;
  emit("search", searchString.value);
}

function handleKeyDown(event: KeyboardEvent) {
  switch (event.key) {
    case "Control":
      // Disabled directly in @keyup
      isCtrlDown.value = true;
      break;
    case "n":
      if (!isCtrlDown.value) break;
    case "ArrowDown":
      if (highlightedIndex.value === null && (props.items?.length || showCreateItemField)) {
        highlightedIndex.value = 0;
      } else if (showCreateItemField.value) {
        highlightedIndex.value = Math.min(
          (highlightedIndex?.value ?? -1) + 1,
          props.items?.length ?? 0,
        );
      } else {
        highlightedIndex.value = Math.min(
          (highlightedIndex?.value ?? -1) + 1,
          (props.items?.length ?? 0) - 1,
        );
      }
      event.preventDefault();
      break;

    case "p":
      if (!isCtrlDown.value) break;
    case "ArrowUp":
      if (highlightedIndex.value === null) {
        break;
      } else if (highlightedIndex.value === 0) {
        highlightedIndex.value = null;
      } else {
        highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0);
      }
      event.preventDefault();
      break;

    case "u":
      if (!isCtrlDown.value) break;
      searchString.value = "";
      highlightedIndex.value = null;
      break;

    case "Enter":
      if (highlightedIndex.value !== null && highlightedIndex.value < (props.items?.length ?? 0)) {
        selectItem(props.items?.[highlightedIndex.value] as T);
      } else if (highlightedIndex.value === props.items?.length) {
        createItem();
      }
      break;

    default:
      break;
  }
}

function selectItem(item: T) {
  emit("select", item);
  searchString.value = "";
}

function createItem() {
  emit("create", searchString.value);
  searchString.value = "";
}

watch(showDropdown, async () => (highlightedIndex.value = null));
</script>

<style scoped>
.search-container {
  width: 100%;
  --max-width: 400px;
  max-width: var(--max-width);
}

.search-dropdown {
  position: absolute;
  background-color: var(--nord0);
  border: 1px solid var(--nord1);
  padding: 0.25em 0.5em;
  z-index: 1;
  display: block;
  width: 100%;
  max-width: var(--max-width);
}

.searching-text {
  color: var(--nord-c1);
}

ul {
  list-style-type: none;
  padding: 0;
  margin: 0;
}

.highlight-item {
  background-color: var(--nord3);
  cursor: pointer;
}

span {
  word-wrap: break-word;
}
</style>
