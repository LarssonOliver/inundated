<template>
  <router-link class="tag-link" :to="to">
    <TagItem :tag="tag" />
  </router-link>
</template>

<script setup lang="ts">
import TagItem from "@/components/tags/TagItem.vue";
import type { Tag } from "@/model";
import type { RouteLocationRaw } from "vue-router";

/** A tag pill that links to the item it stands for, as a list row's primary link. */
defineProps<{
  tag: Tag;
  to: RouteLocationRaw;
}>();
</script>

<style scoped>
/* Neutralizes the global `a` styling (padding, hover background) so the
   pill inside - the only thing providing this link's visuals - isn't
   doubled up with a mismatched box around it. */
.tag-link {
  padding: 0;
  border-radius: var(--radius-md);
  color: inherit;
  transition: opacity var(--transition-fast);
}

.tag-link:hover,
.tag-link:focus {
  background-color: transparent;
  opacity: 0.8;
}

.tag-link:focus-visible {
  box-shadow: var(--focus-ring);
}
</style>
