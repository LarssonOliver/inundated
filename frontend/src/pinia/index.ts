import "pinia";
import type { ErrorState } from "./error-plugin";

declare module "pinia" {
  export interface PiniaCustomProperties {
    $errors: ErrorState;
    $clearError: (action?: string) => void;
  }
}
