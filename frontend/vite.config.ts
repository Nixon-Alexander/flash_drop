import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

// Dev server mem-proxy /api dan /uploads ke backend, jadi tidak perlu CORS saat development.
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const target = env.VITE_API_TARGET || "http://localhost:8080";
  return {
    plugins: [react()],
    server: {
      port: 5173,
      proxy: {
        "/api": { target, changeOrigin: true },
        "/uploads": { target, changeOrigin: true },
      },
      host: true
    },
  };
});
