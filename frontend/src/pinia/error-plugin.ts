import type { PiniaPluginContext } from "pinia";
import { reactive } from "vue";

export interface ActionError {
  action: string;
  error: string;
  timestamp: number;
}

export interface ErrorState {
  last: ActionError | null;
  byAction: Record<string, ActionError[]>;
}

export function errorHandlingPlugin({ store }: PiniaPluginContext) {
  const errors = reactive<ErrorState>({
    last: null,
    byAction: {},
  });

  function recordError(action: string, error: unknown) {
    const entry: ActionError = {
      action,
      error: normalizeError(error),
      timestamp: Date.now(),
    };

    errors.last = entry;
    errors.byAction[action] ??= [];
    errors.byAction[action].push(entry);
  }

  function clearError(action?: string) {
    if (!action) {
      errors.last = null;
      errors.byAction = {};
      return;
    }

    delete errors.byAction[action];
    if (errors.last?.action === action) {
      errors.last = null;
    }
  }

  // Wrap all actions
  store.$onAction(({ name, after, onError }) => {
    onError((error) => {
      recordError(name, error);
    });

    // Optional: clear previous error on success
    after(() => {
      clearError(name);
    });
  });

  return {
    $errors: errors,
    $clearError: clearError,
  };
}

function normalizeError(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Unknown error";
}
