<template>
  <div class="sidebar" :style="sidebarStyle">
    <div v-if="isExpanded" class="sidebar-contianer">
      <div class="title">
        <h2>INUNDATED</h2>
      </div>
      <nav>
        <ul class="link-list">
          <li v-for="route in routes" :key="route.path">
            <router-link :to="route.path">
              {{ route.name || route.path }}
            </router-link>
          </li>
        </ul>
      </nav>
      <div class="footer">
        <strong>
          Inundated v{{ version }}
        </strong>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { version } from "@/../package.json";

const router = useRouter();
const routes = router.getRoutes();

const isExpanded = ref(true);
const toggleSidebar = () => {
  isExpanded.value = !isExpanded.value;
};

const sidebarStyle = computed(() => ({
  width: isExpanded.value ? "200px" : "80px",
}));
</script>

<style scoped>
.sidebar {
  height: 100vh;
  position: sticky;
  padding: 1.5em 2em;
  transition: all 0.1s ease;
  background-color: var(--nord0);
  display: flex;
  flex-direction: column;
}

.sidebar-contianer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar nav {
  flex: 1;
}

.footer strong {
  color: var(--nord3);
}

.title {
  margin-bottom: 1em;
}

.link-list {
  list-style-type: none;
  padding: 0;
}

.link-list li {
  margin: 2em 0;
}

.link-list li a {
  font-size: 1.2em;
}
</style>
