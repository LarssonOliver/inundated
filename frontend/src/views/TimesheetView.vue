<template>
  <div style="display: flex">
    <TagItem :tag="items[0]" />
    <TagItem :tag="items[1]" />
    <TagItem :tag="items[2]" />
  </div>

  <hr />

  <TheTimesheet />
</template>

<script setup lang="ts">
import TagItem from "@/components/tags/TagItem.vue";
import TheTimesheet from "@/components/timesheet/TheTimesheet.vue";
import type { Tag } from "@/model/tag";

const items: Tag[] = [
  { id: 1, name: "tag:test", color: "#44b", userId: 1 },
  { id: 2, name: "tag:hello", color: "#bbb", userId: 1 },
  { id: 3, name: "tag:world", color: "#acd", userId: 1 },
];
</script>
