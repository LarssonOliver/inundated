<template>
  <Teleport to="body">
    <div class="toast-region" role="status" aria-live="polite">
      <Transition name="toast">
        <div
          v-if="history.message"
          :key="history.message.id"
          class="toast"
          :class="{ 'toast--error': history.message.isError }"
        >
          <span class="toast__text">{{ history.message.text }}</span>
          <button
            class="toast__close"
            aria-label="Dismiss"
            title="Dismiss"
            @click="history.dismissMessage()"
          >
            <MaterialIcon icon="close" size="1.1em" />
          </button>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import { useHistoryStore } from "@/stores/history";
import { onUnmounted, watch } from "vue";

/** How long a message stays up, in ms; errors stay twice as long. */
const props = withDefaults(defineProps<{ durationMs?: number }>(), { durationMs: 4000 });

const history = useHistoryStore();
let timer: ReturnType<typeof setTimeout> | undefined;

watch(
  () => history.message?.id,
  () => {
    clearTimeout(timer);
    const message = history.message;
    if (!message) return;
    timer = setTimeout(
      () => {
        if (history.message?.id === message.id) history.dismissMessage();
      },
      message.isError ? props.durationMs * 2 : props.durationMs,
    );
  },
);

onUnmounted(() => clearTimeout(timer));
</script>

<style scoped>
.toast-region {
  position: fixed;
  left: 50%;
  bottom: 1.5em;
  transform: translateX(-50%);
  z-index: 9000;
  max-width: calc(100vw - 2em);
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 0.75em;
  padding: 0.6em 0.6em 0.6em 1em;
  background-color: var(--nord1);
  color: var(--nord6);
  border-left: 3px solid var(--nord8);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  pointer-events: auto;
}

.toast--error {
  border-left-color: var(--nord11);
}

.toast__close {
  display: flex;
  align-items: center;
  width: auto;
  max-width: none;
  padding: 0.25em;
  border: none;
  background: transparent;
  color: var(--nord4);
  border-radius: var(--radius-sm);
}

.toast__close:hover {
  background-color: var(--nord2);
}

.toast-enter-active,
.toast-leave-active {
  transition:
    opacity var(--transition-base),
    transform var(--transition-base);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(0.5em);
}
</style>
