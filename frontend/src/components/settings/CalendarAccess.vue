<template>
  <div class="card calendar-access">
    <h3>Calendar access</h3>
    <p class="note">
      Add a CalDAV account in your calendar app to see your time entries and tasks. Access is
      read-only: edits made in calendar apps are reverted.
    </p>

    <p class="field-label">Server URL</p>
    <div class="copy-row">
      <code>{{ serverUrl }}</code>
      <button class="icon-button" aria-label="Copy server URL" @click="copy(serverUrl)">
        <MaterialIcon :icon="copyIcon(serverUrl)" size="18px" />
      </button>
    </div>

    <p class="field-label">Username</p>
    <p class="note">Anything works. Only the app password is checked.</p>

    <div v-if="created" class="created">
      <p class="field-label">New app password (shown only once)</p>
      <div class="copy-row">
        <code>{{ created.password }}</code>
        <button class="icon-button" aria-label="Copy password" @click="copy(created.password)">
          <MaterialIcon :icon="copyIcon(created.password)" size="18px" />
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
    <p v-else-if="loaded" class="note">No app passwords yet.</p>

    <form class="create-row" @submit.prevent="create">
      <input
        v-model="newName"
        aria-label="App password name"
        placeholder="Name, e.g. iPhone"
        maxlength="100"
      />
      <button type="submit" class="btn-info" :disabled="!newName.trim() || creating">Create</button>
    </form>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <ConfirmationPopup
      :model-value="pendingRevoke !== null"
      title="Revoke app password?"
      :message="
        pendingRevoke
          ? `Calendar apps using &quot;${pendingRevoke.name}&quot; will stop syncing.`
          : undefined
      "
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
const revoking = ref(false);
const loaded = ref(false);
const errorMessage = ref("");
const copiedText = ref("");

function formatDate(date: Date): string {
  return formatDatePickerInput(date, dateFormat.value);
}

function copyIcon(text: string): string {
  return copiedText.value === text ? "check" : "content_copy";
}

async function copy(text: string) {
  errorMessage.value = "";
  try {
    // navigator.clipboard is undefined outside secure contexts (plain HTTP).
    await navigator.clipboard.writeText(text);
    copiedText.value = text;
  } catch {
    copiedText.value = "";
    errorMessage.value = "Could not copy. Select the text and copy it manually.";
  }
}

onMounted(async () => {
  try {
    appPasswords.value = await appPasswordsApi.list();
    loaded.value = true;
  } catch {
    errorMessage.value = "Could not load app passwords.";
  }
});

async function create() {
  creating.value = true;
  errorMessage.value = "";
  try {
    const result = await appPasswordsApi.create(newName.value.trim());
    appPasswords.value.unshift(result.appPassword);
    created.value = { password: result.password };
    newName.value = "";
  } catch {
    errorMessage.value = "Could not create the app password.";
  } finally {
    creating.value = false;
  }
}

async function revoke() {
  const target = pendingRevoke.value;
  if (!target || revoking.value) {
    return;
  }
  revoking.value = true;
  errorMessage.value = "";
  try {
    await appPasswordsApi.delete(target.id);
    appPasswords.value = appPasswords.value.filter((appPassword) => appPassword.id !== target.id);
  } catch {
    errorMessage.value = "Could not revoke the app password.";
  } finally {
    revoking.value = false;
  }
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
  padding-bottom: 0.5em;
}

.copy-row code {
  flex: 1;
  overflow-wrap: anywhere;
}

.error {
  color: var(--nord11);
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
  display: inline-flex;
  width: auto;
  max-width: none;
  background: none;
  padding: 0.25em;
  color: inherit;
}

.create-row input {
  flex: 1;
}

.create-row button {
  width: auto;
}
</style>
