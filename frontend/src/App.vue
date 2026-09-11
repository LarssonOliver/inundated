<script setup lang="ts">
import { onMounted } from "vue";
import { RouterView } from "vue-router";
import TheMenu from "./components/TheMenu.vue";
import LoadingScreen from "./components/LoadingScreen.vue";
import { useUserStore } from "@/stores/user";
import { useStartup, runStartupProbe } from "@/composables/useStartup";

const userStore = useUserStore();
const { isStarting } = useStartup();

onMounted(() => {
  void runStartupProbe(userStore.fetchCurrentUser());
});
</script>

<template>
  <LoadingScreen v-if="isStarting" />
  <template v-else>
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
  padding: 1em 1.5em;
  overflow-y: scroll;
}
</style>
