/// <reference types="vitest/config" />
import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueDevTools from "vite-plugin-vue-devtools";
import Components from "unplugin-vue-components/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools(), Components({})],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  base: "/",
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        // Keep the browser's Host header so it matches the Origin the server's
        // CSRF check compares it against.
        changeOrigin: false,
        secure: false,
      },
    },
  },
  test: {
    environment: "jsdom",
    // Node 22+ ships a native, opt-in `localStorage` global that stays inert and
    // prints an ExperimentalWarning unless started with --localstorage-file.
    // Turning it off lets jsdom install its own working localStorage instead.
    execArgv: ["--no-experimental-webstorage"],
  },
});
