<template>
  <div>
    <input
      type="text"
      :placeholder="props.placeholder"
      @mouseenter="showDropdown = true"
      @mouseleave="showDropdown = false"
    />

    <div
      v-if="showDropdown"
      class="dropdown"
      @mouseenter="showDropdown = true"
      @mouseleave="showDropdown = false"
    >
      <div v-if="searching" class="searching-text">Searching...</div>
      <div v-else></div>

      <slot />
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import { ref } from "vue";

const props = defineProps<{
  items?: T[];
  searchFunc?: (query: string) => T[];
  placeholder?: string;
}>();

const showDropdown = ref(false);
const searching = ref(true);
</script>

<style scoped>
.dropdown {
  position: absolute;
  background-color: var(--nord0);
  border: 1px solid var(--nord1);
  padding: 0.5em 0.75em;
  z-index: 1;
  display: block;
  width: 100%;
  max-width: 398px;
}

.searching-text {
  color: var(--nord-c1);
}
</style>
