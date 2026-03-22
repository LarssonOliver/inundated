<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="modelValue"
        class="overlay"
        @click.self="onCancel"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="descId"
      >
        <Transition name="popup">
          <div v-if="modelValue" class="popup">
            <!-- Content -->
            <div class="popup__body">
              <h2 :id="titleId" class="popup__title">{{ title }}</h2>
              <p :id="descId" class="popup__message">{{ message }}</p>
            </div>

            <!-- Actions -->
            <div class="popup__actions">
              <button class="btn btn--ghost" @click="onCancel" ref="cancelRef">
                {{ cancelLabel }}
              </button>
              <button class="btn" :class="`btn-${variant}`" @click="onConfirm">
                {{ confirmLabel }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { generateSecureUid } from "@/helpers/uid";
import { ref, watch, onMounted, onUnmounted, computed } from "vue";

// ─── Types ────────────────────────────────────────────────────────────────────

type Variant = "error" | "danger" | "warning" | "success" | "info";

// ─── Props ────────────────────────────────────────────────────────────────────

interface Props {
  /** Controls visibility via v-model */
  modelValue: boolean;
  title?: string;
  message?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Visual variant controlling icon and button colour */
  variant?: Variant;
  /** Close when clicking the backdrop */
  closeOnBackdrop?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  title: "Are you sure?",
  message: "This action cannot be undone.",
  confirmLabel: "Confirm",
  cancelLabel: "Cancel",
  variant: "info",
  closeOnBackdrop: true,
});

// ─── Emits ────────────────────────────────────────────────────────────────────

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "confirm"): void;
  (e: "cancel"): void;
}>();

// ─── Refs ─────────────────────────────────────────────────────────────────────

const cancelRef = ref<HTMLButtonElement | null>(null);

// Unique IDs for ARIA
const uid = generateSecureUid(6);
const titleId = computed(() => `popup-title-${uid}`);
const descId = computed(() => `popup-desc-${uid}`);

// ─── Handlers ─────────────────────────────────────────────────────────────────

function onConfirm(): void {
  emit("confirm");
  emit("update:modelValue", false);
}

function onCancel(): void {
  if (!props.closeOnBackdrop) return;
  emit("cancel");
  emit("update:modelValue", false);
}

// ─── Keyboard trap ────────────────────────────────────────────────────────────

function handleKeydown(e: KeyboardEvent): void {
  if (!props.modelValue) return;
  if (e.key === "Escape") onCancel();
}

onMounted(() => window.addEventListener("keydown", handleKeydown));
onUnmounted(() => window.removeEventListener("keydown", handleKeydown));

// Focus the cancel button when the dialog opens
watch(
  () => props.modelValue,
  (val: boolean) => {
    if (val) {
      // Wait for the transition to finish before focusing
      setTimeout(() => cancelRef.value?.focus(), 80);
    }
  },
);
</script>

<style scoped>
/* ── Variables ─────────────────────────────────────────────────────────────── */
.overlay {
  --transition: 240ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* ── Backdrop ───────────────────────────────────────────────────────────────── */
.overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  backdrop-filter: blur(3px);
  -webkit-backdrop-filter: blur(3px);
}

/* ── Popup card ─────────────────────────────────────────────────────────────── */
.popup {
  background: var(--nord-c0);
  border-radius: 4px;
  box-shadow:
    0 20px 60px -10px rgba(0, 0, 0, 0.18),
    0 4px 16px -4px rgba(0, 0, 0, 0.12);
  border: 1px solid var(--nord0);
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  outline: none;
}

/* ── Body ───────────────────────────────────────────────────────────────────── */
.popup__body {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.popup__title {
  margin: 0;
}

.popup__message {
  margin: 0;
}

/* ── Actions ────────────────────────────────────────────────────────────────── */
.popup__actions {
  display: flex;
  gap: 0.625rem;
  justify-content: flex-end;
  padding-top: 0.25rem;
}
</style>
