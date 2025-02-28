<template>
  <div>
    <DataTable size="small" :value="items">
      <Column field="test1" header="Test"></Column>
      <Column field="test2" header="Test"></Column>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const items = ref<{ test1: string, test2: string }[]>([
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
  {
    test2: "test2",
    test1: "test1"
  },
]);

</script>
