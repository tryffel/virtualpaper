import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { VitePWA } from "vite-plugin-pwa";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      devOptions: {
        enabled: true,
      },
      manifestFilename: "manifest.json",
      manifest: {
        name: "Virtualpaper",
        background_color: "#313131",
        theme_color: "#673ab7",
        display: "standalone",
        scope: "/",
        start_url: "./index.html",
        icons: [
          {
            src: "logo192.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "any",
          },
          {
            src: "favicon.ico",
            sizes: "16x16",
            type: "image/x-icon",
            purpose: "any",
          },
          {
            src: "favicon-16x16.png",
            sizes: "16x16",
            type: "image/png",
          },
          {
            src: "favicon-32x32.png",
            sizes: "32x32",
            type: "image/png",
          },
        ],
      },
    }),
  ],
  define: {
    "process.env": process.env,
  },
});
