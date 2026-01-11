<template>
  <div class="sidebar" :style="sidebarStyle" @click="toggleSidebar">Testar</div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const isExpanded = ref(false);
const toggleSidebar = () => {
  isExpanded.value = !isExpanded.value;
};

const sidebarStyle = computed(() => ({
  width: isExpanded.value ? "250px" : "80px",
}));
</script>

<style scoped>
.sidebar {
  outline: 1px solid #ccc;
  height: 100vh;
  position: sticky;
  margin-right: 1em;
  padding: 0.25em 0.5em;
  transition: all 0.1s ease;
}
</style>
