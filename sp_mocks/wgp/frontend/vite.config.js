import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { VitePWA } from 'vite-plugin-pwa'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      devOptions: {
        enabled: true
      },
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'mask-icon.svg'],
      manifest: {
        name: "We've Got Poems",
        short_name: "WGP",
        theme_color: "#fff",
        description: "We've got poems by Layer8.",
        icons: [
          {
            src: "icon-48.png",
            sizes: "48x48",
            type: "image/png"
          },
          {
            src: "icon-96.png",
            sizes: "96x96",
            type: "image/png"
          },
          {
            src: "icon-192.png",
            sizes: "192x192",
            type: "image/png"
          },
          {
            src: "icon-512.png",
            sizes: "512x512",
            type: "image/png"
          },
          {
            src: "icon-750.png",
            sizes: "750x750",
            type: "image/png"
          }
        ]
      }
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
