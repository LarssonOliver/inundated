<template>
  <div>
    <Menubar :model="items">
      <template #item="{ item, props }">
        <a v-ripple class="flex items-center" v-bind="props.action">
          <span v-if="item.icon" :class="item.icon"></span>
          <span>{{ item.label }}</span>
        </a>
      </template>
      <template #end>
        <Avatar icon="pi pi-user" shape="circle" size="normal" />
      </template>
    </Menubar>
  </div>
</template>


<script setup lang="ts">
import type { MenuItem } from "primevue/menuitem";
import { ref } from "vue";

const items = ref<MenuItem[]>([
  {
    label: 'Home',
    icon: 'pi pi-home'
  },
]);

</script>
