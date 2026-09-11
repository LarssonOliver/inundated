<template>
  <div class="sidebar" :style="sidebarStyle">
    <div class="sidebar-container">
      <div class="title">
        <h2 v-if="isExpanded">INUNDATED</h2>
        <h2 v-else class="title-mark">I</h2>
        <button
          class="icon-button toggle-button"
          @click="toggleSidebar"
          :aria-expanded="isExpanded"
          :title="isExpanded ? 'Collapse sidebar' : 'Expand sidebar'"
        >
          <MaterialIcon :icon="isExpanded ? 'chevron_left' : 'chevron_right'" size="1.2em" />
        </button>
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
        <div v-if="userStore.user" class="account">
          <span v-if="isExpanded" class="account-name" :title="userStore.user.email">
            {{ userStore.user.name || userStore.user.email }}
          </span>
          <button
            class="btn-info logout"
            :class="{ 'logout-compact': !isExpanded }"
            :title="!isExpanded ? 'Log out' : ''"
            @click="userStore.logout()"
          >
            <MaterialIcon v-if="!isExpanded" icon="logout" size="1.1em" />
            <span v-else>Log out</span>
          </button>
        </div>
        <strong v-if="isExpanded"> Inundated {{ version }} </strong>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { version } from "@/../package.json";
import { useUserStore } from "@/stores/user";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";

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
  width: isExpanded.value ? "220px" : "76px",
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

.account {
  display: flex;
  flex-direction: column;
  gap: 0.5em;
  margin-bottom: 0.75em;
}

.account-name {
  color: var(--nord4);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.logout {
  width: 100%;
}

.logout-compact {
  width: 44px;
  padding: 0.5em;
  display: flex;
  align-items: center;
  justify-content: center;
}

.title {
  margin-bottom: 1.5em;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 2em;
}

.title h2 {
  margin: 0;
  font-size: 1.2em;
  white-space: nowrap;
  overflow: hidden;
}

.title-mark {
  color: var(--nord8);
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
</style>
