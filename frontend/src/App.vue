<script setup lang="ts">
import { RouterView } from "vue-router";
import TheMenu from "./components/TheMenu.vue";
</script>

<template>
  <header></header>

  <TheMenu />

  <main>
    <RouterView />
  </main>
</template>

<style scoped></style>
