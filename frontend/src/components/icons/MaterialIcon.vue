<template>
  <span class="material-icons">{{ props.icon }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  icon: string;
  size?: string;
  color?: string;
}>();

const iconSize = computed(() => {
  return props.size ? props.size : "24px";
});

const iconColor = computed(() => {
  return props.color ? props.color : "inherit";
});
</script>

<style scoped>
span {
  font-size: v-bind(iconSize);
  color: v-bind(iconColor);
}
</style>
