<template>
  <div class="card calendar-access">
    <h3>Calendar access</h3>
    <p class="note">
      Subscribe from a calendar app to see your tasks. Access is read-only: Edits made in calendar
      apps are reverted.
    </p>

    <p class="field-label">Server URL</p>
    <div class="copy-row">
      <code>{{ serverUrl }}</code>
      <button class="icon-button" aria-label="Copy server URL" @click="copy(serverUrl)">
        <MaterialIcon icon="content_copy" size="18px" />
      </button>
    </div>

    <p class="field-label">Username</p>
    <p class="note">Anything works. Only the app password is checked.</p>

    <div v-if="created" class="created">
      <p class="field-label">New app password (shown only once)</p>
      <div class="copy-row">
        <code>{{ created.password }}</code>
        <button class="icon-button" aria-label="Copy password" @click="copy(created.password)">
          <MaterialIcon icon="content_copy" size="18px" />
        </button>
      </div>
      <button class="btn-info" @click="created = null">Done</button>
    </div>

    <p class="field-label">App passwords</p>
    <ul v-if="appPasswords.length > 0" class="app-passwords">
      <li v-for="appPassword in appPasswords" :key="appPassword.id">
        <div class="app-password-details">
          <span class="app-password-name">{{ appPassword.name }}</span>
          <span class="note">
            Created {{ formatDate(appPassword.createdAt) }} ·
            {{
              appPassword.lastUsedAt
                ? `last used ${formatDate(appPassword.lastUsedAt)}`
                : "never used"
            }}
          </span>
        </div>
        <button
          class="icon-button"
          :aria-label="`Revoke ${appPassword.name}`"
          @click="pendingRevoke = appPassword"
        >
          <MaterialIcon icon="delete" size="18px" />
        </button>
      </li>
    </ul>
    <p v-else class="note">No app passwords yet.</p>

    <form class="create-row" @submit.prevent="create">
      <input v-model="newName" placeholder="Name, e.g. iPhone" maxlength="100" />
      <button type="submit" class="btn-info" :disabled="!newName.trim() || creating">Create</button>
    </form>

    <ConfirmationPopup
      :model-value="pendingRevoke !== null"
      title="Revoke app password?"
      :message="`Calendar apps using &quot;${pendingRevoke?.name}&quot; will stop syncing.`"
      confirm-label="Revoke"
      variant="danger"
      @update:model-value="pendingRevoke = null"
      @confirm="revoke"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { appPasswordsApi } from "@/api";
import type { AppPassword } from "@/model";
import { useSettingsStore } from "@/stores/settings";
import { formatDatePickerInput } from "@/helpers/dates";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";

const serverUrl = `${window.location.origin}/dav/me/`;

const settingsStore = useSettingsStore();
const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");

const appPasswords = ref<AppPassword[]>([]);
const newName = ref("");
const creating = ref(false);
const created = ref<{ password: string } | null>(null);
const pendingRevoke = ref<AppPassword | null>(null);

function formatDate(date: Date): string {
  return formatDatePickerInput(date, dateFormat.value);
}

function copy(text: string) {
  navigator.clipboard.writeText(text);
}

onMounted(async () => {
  appPasswords.value = await appPasswordsApi.list();
});

async function create() {
  creating.value = true;
  try {
    const result = await appPasswordsApi.create(newName.value.trim());
    appPasswords.value.push(result.appPassword);
    created.value = { password: result.password };
    newName.value = "";
  } finally {
    creating.value = false;
  }
}

async function revoke() {
  const target = pendingRevoke.value;
  if (!target) {
    return;
  }
  await appPasswordsApi.delete(target.id);
  appPasswords.value = appPasswords.value.filter((appPassword) => appPassword.id !== target.id);
}
</script>

<style scoped>
.calendar-access {
  max-width: 400px;
  display: flex;
  flex-direction: column;
}

.card {
  background-color: var(--nord0);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: 1.5em;
  padding-top: 0.5em;
}

.field-label {
  margin-bottom: 0.25em;
  color: var(--nord3);
}

.note {
  margin: 0 0 0.5em;
  color: var(--nord4);
  font-size: 0.9em;
}

.copy-row,
.create-row,
.app-passwords li {
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.copy-row code {
  flex: 1;
  overflow-wrap: anywhere;
}

.created {
  margin: 1em 0;
}

.app-passwords {
  list-style: none;
  padding: 0;
  margin: 0 0 1em;
}

.app-password-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.icon-button {
  background: none;
  padding: 0.25em;
  color: inherit;
}

.create-row input {
  flex: 1;
}
</style>
