import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/auth": "http://localhost:8080",
      "/users": "http://localhost:8080",
      "/solar-profiles": "http://localhost:8080",
      "/forecast": "http://localhost:8080",
      "/devices": "http://localhost:8080",
      "/notifications": "http://localhost:8080",
      "/health": "http://localhost:8080",
    },
  },
});
