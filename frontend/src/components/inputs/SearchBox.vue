<template>
  <div class="container">
    <input type="text" :placeholder="props.placeholder" @mouseenter="showDropdown = true"
      @mouseleave="showDropdown = true" />

    <div v-if="showDropdown" class="dropdown" @mouseenter="showDropdown = true" @mouseleave="showDropdown = true">
      <div v-if="searching" class="searching-text">Searching...</div>
      <ul v-else>
        <li v-for="(item, _) of items">
          <div class="list-item">
            <slot v-bind="item" />
          </div>
          <hr />
        </li>
        <li>
          <div class="list-item">
            Create new item...
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import { ref } from "vue";

const items = ref<T[]>([
  { name: "test" },
  { name: "test2" },
  { name: "test3" },
]);

const props = defineProps<{
  // items?: T[];
  placeholder?: string;
  searchFunc?: (query: string) => T[];
  createFunc?: (query: string) => Promise<T>;
}>();

const showDropdown = ref(true);
const searching = ref(false);
</script>

<style scoped>
.container {
  --max-width: 400px;
  max-width: var(--max-width);
}

.dropdown {
  position: absolute;
  background-color: var(--nord0);
  border: 1px solid var(--nord1);
  padding: 0.5em 0.75em;
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

.list-item:hover {
  background-color: var(--nord3);
}
</style>
