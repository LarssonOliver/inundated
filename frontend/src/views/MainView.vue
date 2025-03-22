<template>
  <div style="display: flex">
    <TagItem color="green" name="tag:test" />
    <TagItem color="black" name="tag:test" />
    <TagItem color="#fff" name="tag:test" />
  </div>

  <hr />

  <SearchBox placeholder="test...">
    <template v-slot="item">
      {{ item }}
    </template>
  </SearchBox>
</template>

<script setup lang="ts">
import SearchBox from "@/components/inputs/SearchBox.vue";
import TagItem from "@/components/tags/TagItem.vue";
// import TagSearchOrCreateBox from "@/components/tags/TagSearchOrCreateBox.vue";
</script>
