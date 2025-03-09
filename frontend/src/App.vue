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

<style>
@import "@/../node_modules/@fontsource/roboto/index.css";
@import "@/../node_modules/material-icons/iconfont/material-icons.css";
@import "@/assets/nord.css";
@import "@/assets/main.css";
</style>
