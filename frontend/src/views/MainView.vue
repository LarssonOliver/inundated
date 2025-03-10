<template>
  <div style="display: flex">
    <TagItem color="green" name="tag:test" />
    <TagItem color="black" name="tag:test" />
    <TagItem color="#fff" name="tag:test" />
  </div>
  <!-- <TagSearchOrCreateBox /> -->
  <div style="display: block">
    <input type="text" placeholder="testing" value="test" />

    <input type="button" value="test" />
    <input type="checkbox" />
    <input type="color" />
    <input type="date" />
    <input type="datetime-local" />
    <input type="email" />
    <input type="file" />
    <input type="hidden" />
    <input type="image" />
    <input type="month" />
    <input type="number" />
    <input type="password" />
    <input type="radio" />
    <input type="range" />
    <input type="reset" />
    <input type="search" />
    <input type="submit" />
    <input type="tel" />
    <input type="text" required />
    <input type="time" />
    <input type="url" />
    <input type="week" />
  </div>
</template>

<script setup lang="ts">
import TagItem from "@/components/tags/TagItem.vue";
// import TagSearchOrCreateBox from "@/components/tags/TagSearchOrCreateBox.vue";
</script>
