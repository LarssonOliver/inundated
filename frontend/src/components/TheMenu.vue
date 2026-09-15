<template>
  <div class="sidebar" :style="sidebarStyle">
    <div class="sidebar-container">
      <div class="title">
        <img src="/logo.svg" alt="Inundated" class="brand-mark" />
        <h2 v-if="isExpanded" class="brand-name">Inundated</h2>
      </div>
      <nav>
        <ul class="link-list">
          <li v-for="link in links" :key="link.to">
            <router-link :to="link.to" class="nav-link" :title="!isExpanded ? link.label : ''">
              <MaterialIcon class="nav-icon" :icon="link.icon" size="1.3em" />
              <span v-if="isExpanded" class="nav-label">{{ link.label }}</span>
            </router-link>
          </li>
        </ul>
      </nav>
      <div class="footer">
        <router-link
          to="/settings"
          class="nav-link settings-link"
          :title="!isExpanded ? 'Settings' : ''"
        >
          <MaterialIcon
            class="nav-icon"
            :icon="userStore.user ? 'account_circle' : 'settings'"
            size="1.3em"
          />
          <span v-if="isExpanded" class="nav-label">{{
            userStore.user ? userStore.user.name || userStore.user.email : "Settings"
          }}</span>
        </router-link>
        <div class="bottom">
          <strong v-if="isExpanded"> Inundated {{ version }} </strong>
          <button
            class="icon-button"
            :class="{ 'toggle-button-compact': !isExpanded }"
            @click="toggleSidebar"
            :aria-expanded="isExpanded"
            :title="isExpanded ? 'Collapse sidebar' : 'Expand sidebar'"
          >
            <MaterialIcon :icon="isExpanded ? 'chevron_left' : 'chevron_right'" size="1.2em" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { version } from "@/../package.json";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import { useUserStore } from "@/stores/user";

const userStore = useUserStore();

const isExpanded = ref(true);
const toggleSidebar = () => {
  isExpanded.value = !isExpanded.value;
};

const links = [
  { to: "/", label: "Timesheet", icon: "schedule" },
  { to: "/projects", label: "Projects", icon: "folder" },
  { to: "/tags", label: "Tags", icon: "sell" },
];

const sidebarStyle = computed(() => ({
  width: isExpanded.value ? "220px" : "65px",
}));
</script>

<style scoped>
.sidebar {
  height: 100vh;
  position: sticky;
  padding: 1.25em 1em;
  transition: width var(--transition-base);
  background-color: var(--nord0);
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  z-index: 1;
}

.sidebar-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.sidebar nav {
  flex: 1;
}

.footer strong {
  color: var(--nord3);
  font-size: 0.85em;
  white-space: nowrap;
}

.settings-link {
  margin-bottom: 0.75em;
}

.title {
  margin-bottom: 1.5em;
  display: flex;
  align-items: center;
  gap: 0.6em;
  min-height: 2em;
}

.brand-mark {
  width: 3em;
  height: 3em;
  flex-shrink: 0;
}

.brand-name {
  margin: 0;
  font-family: "PT Sans", sans-serif;
  font-weight: 700;
  font-size: 2em;
  color: var(--nord6);
  white-space: nowrap;
  overflow: hidden;
}

.bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.icon-button {
  width: auto;
  max-width: none;
  padding: 0.4em;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--nord4);
  flex-shrink: 0;
}

.icon-button:hover {
  background-color: var(--nord1);
}

.link-list {
  list-style-type: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4em;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75em;
  padding: 0.6em 0.75em;
  border-radius: var(--radius-sm);
  border-left: 3px solid transparent;
  color: var(--nord4);
  transition:
    background-color var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.nav-link:hover {
  background-color: var(--nord1);
}

.nav-icon {
  flex-shrink: 0;
}

.nav-label {
  font-size: 1.05em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-link.router-link-active {
  background-color: var(--nord1);
  border-left-color: var(--nord8);
  color: var(--nord6);
}

.toggle-button-compact {
  width: 100%;
}
</style>
