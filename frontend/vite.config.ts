import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// O dev server faz proxy de /api para o backend, evitando CORS em
// desenvolvimento e permitindo que o cookie de sessão seja tratado como
// same-origin.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
