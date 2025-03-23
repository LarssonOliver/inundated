<template>
  <div :class="containerClasses" @mouseenter="hover = true" @mouseleave="hover = false">
    <div>{{ tag.name }}</div>
    <MaterialIcon
      @click="(onClose as () => void)()"
      v-if="hover && onClose !== undefined"
      class="close-icon"
      icon="close"
      size="20px"
    />
  </div>
</template>

<script setup lang="ts">
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import { shouldTextBeDarkFromBgColor } from "@/helpers/colors";
import type { Tag } from "@/model/model";
import { reactive, ref } from "vue";

const hover = ref(false);

const { tag, onClose } = defineProps<{
  tag: Tag;
  onClose?: () => void;
}>();

const darkText = shouldTextBeDarkFromBgColor(tag.color);
const containerClasses = reactive({
  "tag-container": true,
  "dark-text": darkText,
});
</script>

<style scoped>
.tag-container {
  background-color: v-bind("tag.color");
  margin: 0.25em 0.25em 0.25em 0;
  padding: 0.25em 0.5em;
  width: fit-content;
  border-radius: 0.5em;
  border: 1px solid var(--nord1);
  display: flex;
}

.tag-container.dark-text {
  color: var(--nord0);
}

.close-icon {
  margin-left: 0.25em;
  cursor: pointer;
}
</style>
