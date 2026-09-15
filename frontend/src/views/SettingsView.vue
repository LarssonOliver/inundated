<template>
  <div class="settings-page">
    <div class="title-bar">
      <h2>Settings</h2>
    </div>

    <div v-if="userStore.user" class="card account-card">
      <div class="account-info-field" v-if="userStore.user.name">
        <p class="field-label">Name</p>
        <span class="account-name" :title="userStore.user.email">
          {{ userStore.user.name }}
        </span>
      </div>
      <div class="account-info-field" v-if="userStore.user.email">
        <p class="field-label">Email</p>
        <span class="account-name" :title="userStore.user.email">
          {{ userStore.user.email }}
        </span>
      </div>
      <button class="btn-info logout" @click="userStore.logout()">Log out</button>
    </div>

    <div v-if="model" class="card settings-edit">
      <p class="field-label">Start of Week</p>
      <Dropdown v-model="model.weekStartDay" :options="weekStartDayOptions" />

      <p class="field-label">Timezone</p>
      <Dropdown
        v-model="model.timezone"
        :options="timezoneOptions"
        searchable
        placeholder="Search timezones..."
      />

      <p class="field-label">Duration Format</p>
      <Dropdown v-model="model.durationFormat" :options="durationFormatOptions" />

      <p class="field-label">Time Format</p>
      <Dropdown v-model="model.timeFormat" :options="timeFormatOptions" />

      <p class="field-label">Date Format</p>
      <Dropdown v-model="model.dateFormat" :options="dateFormatOptions" />

      <div class="button-container">
        <button class="btn-info" :disabled="saving" @click="save">
          {{ saving ? "Saving..." : "Save" }}
        </button>
        <span v-if="justSaved" class="saved-hint">Saved</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useSettingsStore } from "@/stores/settings";
import { useUserStore } from "@/stores/user";
import type { Settings } from "@/model";
import Dropdown, { type DropdownOption } from "@/components/inputs/SelectDropdown.vue";
import { timezoneOptions } from "@/helpers/timezones";
import { diffSettings } from "@/helpers/settingsDiff";

const weekStartDayOptions: DropdownOption[] = [
  { value: "monday", label: "Monday" },
  { value: "sunday", label: "Sunday" },
];

const durationFormatOptions: DropdownOption[] = [
  { value: "long", label: "Long (2h 30m)" },
  { value: "decimal", label: "Decimal (2.5h)" },
  { value: "clock", label: "Clock (02:30)" },
];

const timeFormatOptions: DropdownOption[] = [
  { value: "24h", label: "24-hour" },
  { value: "12h", label: "12-hour" },
];

const dateFormatOptions: DropdownOption[] = [
  { value: "iso", label: "2024-01-15" },
  { value: "us", label: "01/15/2024" },
  { value: "eu", label: "15/01/2024" },
  { value: "text", label: "15 Jan 2024" },
];

const settingsStore = useSettingsStore();
const userStore = useUserStore();

const model = ref<Settings | null>(null);
// The snapshot save() diffs against, so a save only PATCHes fields this page
// actually changed - not the whole object, which would risk silently
// clobbering a change another tab saved in the meantime.
const original = ref<Settings | null>(null);
const saving = ref(false);
const justSaved = ref(false);

onMounted(async () => {
  if (!settingsStore.settings) {
    await settingsStore.fetchSettings();
  }
});

watch(
  () => settingsStore.settings,
  (loaded) => {
    if (loaded && !model.value) {
      model.value = { ...loaded };
      original.value = { ...loaded };
    }
  },
  { immediate: true },
);

async function save() {
  if (!model.value || !original.value) {
    return;
  }

  const patch = diffSettings(original.value, model.value);
  if (Object.keys(patch).length === 0) {
    return;
  }

  saving.value = true;
  justSaved.value = false;
  try {
    await settingsStore.updateSettings(patch);
    if (settingsStore.settings) {
      model.value = { ...settingsStore.settings };
      original.value = { ...settingsStore.settings };
    }
    justSaved.value = true;
  } finally {
    saving.value = false;
  }
}
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 1.5em;
}

.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: -1em;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
  padding: 0.25em 0;
}

.card {
  background-color: var(--nord0);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: 1.5em;
  padding-top: 0.5em;
}

.account-card,
.settings-edit {
  max-width: 400px;
}

.account-info-field {
  margin-bottom: 1em;
}

.account-name {
  color: var(--nord4);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.settings-edit {
  display: flex;
  flex-direction: column;
}

.field-label {
  margin-bottom: 0.25em;
  color: var(--nord3);
}

.settings-edit > :deep(.dropdown) {
  margin-bottom: 1em;
}

.button-container {
  margin-top: 1em;
  display: flex;
  align-items: center;
  gap: 1em;
}

.saved-hint {
  color: var(--nord14);
}

.logout {
  margin-top: 1em;
}
</style>
