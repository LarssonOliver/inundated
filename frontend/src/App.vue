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
  display: grid;
  grid-template-columns: auto 1fr;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  top: 0;
  position: sticky;
  overflow-y: auto;
}

.content {
  flex: 1;
  padding: 0.25em 0.5em;
  overflow-y: scroll;
}
</style>
