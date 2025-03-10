<template>
  <div>
    <input
      type="text"
      v-model="searchQuery"
      @input="filterItems"
      placeholder="Search or create..."
    />
    <ul v-if="filteredItems.length > 0">
      <li v-for="item in filteredItems" :key="item" @click="selectItem(item)">
        {{ item }}
      </li>
    </ul>
    <div v-else @click="createNewItem">Create new item: "{{ searchQuery }}"</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const items = ref<string[]>(["Item 1", "Item 2", "Item 3"]);
const searchQuery = ref("");
const filteredItems = ref<string[]>([]);

const filterItems = () => {
  filteredItems.value = items.value.filter((item) =>
    item.toLowerCase().includes(searchQuery.value.toLowerCase()),
  );
};

const selectItem = (item: string) => {
  searchQuery.value = item;
  filteredItems.value = [];
};

const createNewItem = () => {
  if (searchQuery.value && !items.value.includes(searchQuery.value)) {
    items.value.push(searchQuery.value);
    searchQuery.value = "";
    filteredItems.value = [];
  }
};
</script>

<style scoped>
input {
  width: 100%;
  padding: 8px;
  margin-bottom: 8px;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  padding: 8px;
  cursor: pointer;
}

div {
  padding: 8px;
  cursor: pointer;
  color: blue;
}
</style>
