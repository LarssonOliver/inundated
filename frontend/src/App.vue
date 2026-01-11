<script setup lang="ts">
import { RouterView } from "vue-router";
import TheMenu from "./components/TheMenu.vue";
</script>

<template>
  <header></header>

  <div class="layout">
    <aside class="sidebar">
      <TheMenu />
    </aside>

    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style>
@import "@/../node_modules/@fontsource/roboto/index.css";
@import "@/../node_modules/material-icons/iconfont/material-icons.css";
@import "@/assets/nord.css";
@import "@/assets/main.css";
</style>

<style scoped>
.layout {
  top: 0;
  left: 0;
  display: flex;
  min-height: 100vh;
}

.sidebar {
  flex-shrink: 0;
}

.content {
  flex: 1;
}
</style>
