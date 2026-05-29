<template>
  <div
    class="skeleton"
    :class="[`skeleton--${variant}`, { 'skeleton--animated': animated }]"
    :style="skeletonStyle"
    role="status"
    aria-label="Loading..."
    aria-busy="true"
  />
</template>

<script setup lang="ts">
import { computed, type CSSProperties } from "vue";

type SkeletonVariant = "text" | "circular" | "rectangular" | "rounded";

interface Props {
  variant?: SkeletonVariant;
  width?: string | number;
  height?: string | number;
  animated?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "text",
  width: "100%",
  height: undefined,
  animated: true,
});

const toUnit = (value: string | number | undefined): string | undefined => {
  if (value === undefined) return undefined;
  return typeof value === "number" ? `${value}px` : value;
};

const skeletonStyle = computed<CSSProperties>(() => ({
  width: toUnit(props.width),
  height: toUnit(props.height),
}));
</script>

<style scoped>
.skeleton {
  display: block;
  background-color: var(--nord0);
  position: relative;
  overflow: hidden;
}

/* Variant: text — mimics a line of text */
.skeleton--text {
  height: 1em;
  border-radius: 4px;
  transform: scale(1, 0.8);
  transform-origin: 0 50%;
}

/* Variant: circular — avatar / icon placeholder */
.skeleton--circular {
  border-radius: 50%;
  width: 40px;
  height: 40px;
}

/* Variant: rectangular — image / card placeholder */
.skeleton--rectangular {
  border-radius: 0;
}

/* Variant: rounded — softer card placeholder */
.skeleton--rounded {
  border-radius: 8px;
}

/* Shimmer animation */
.skeleton--animated::after {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(216, 222, 233, 0.08) 50%,
    transparent 100%
  );
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.6s ease-in-out infinite;
}

@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }

  100% {
    background-position: -200% 0;
  }
}

/* Reduced-motion: respect user preference */
@media (prefers-reduced-motion: reduce) {
  .skeleton--animated::after {
    animation: none;
  }
}
</style>
