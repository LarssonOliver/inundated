<template>
  <div class="loading-screen" role="status" aria-label="Loading">
    <div class="spinner"></div>
  </div>
</template>

<style scoped>
.loading-screen {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--nord-c0);
  z-index: 1000;
}

.spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid var(--nord3);
  border-top-color: var(--nord8);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .spinner {
    animation-duration: 2.4s;
  }
}
</style>
