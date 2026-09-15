<template>
  <div class="settings-page">
    <div class="title-bar">
      <h2>Settings</h2>
    </div>
    <div v-if="model" class="card settings-edit">
      <p class="field-label">Week starts on</p>
      <select v-model="model.weekStartDay">
        <option value="monday">Monday</option>
        <option value="sunday">Sunday</option>
      </select>

      <p class="field-label">Timezone</p>
      <input v-model="model.timezone" type="text" placeholder="e.g. Europe/Stockholm or UTC" />

      <p class="field-label">Duration format</p>
      <select v-model="model.durationFormat">
        <option value="long">Long (2h 30m)</option>
        <option value="decimal">Decimal (2.5h)</option>
        <option value="clock">Clock (02:30)</option>
      </select>

      <p class="field-label">Time format</p>
      <select v-model="model.timeFormat">
        <option value="24h">24-hour</option>
        <option value="12h">12-hour</option>
      </select>

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
import type { Settings } from "@/model";

const settingsStore = useSettingsStore();

const model = ref<Settings | null>(null);
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
    }
  },
  { immediate: true },
);

async function save() {
  if (!model.value) {
    return;
  }

  saving.value = true;
  justSaved.value = false;
  try {
    await settingsStore.updateSettings({ ...model.value });
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
}

.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: 1.25em;
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

.settings-edit {
  display: flex;
  flex-direction: column;
  max-width: 400px;
}

.field-label {
  margin-bottom: 0.25em;
  color: var(--nord3);
}

select {
  width: 100%;
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
</style>
