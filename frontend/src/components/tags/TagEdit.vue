<template>
  <div class="tag-edit"></div>
</template>

<script setup lang="ts">
import type { Tag } from "@/model";

const model = defineModel<Tag>();
</script>

<style scoped>
.tag-edit {
  margin: 1em 1em;
  display: flex;
  flex-direction: column;
  border: 1px solid red;
  max-width: 400px;
}
</style>
