import { resolve } from "node:path";
import { defineConfig, loadEnv } from "vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import viteReact from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig(({ mode }) => {
  // The deployed site's URL, from `pnpm web:env`: `vite dev` sends the
  // catalogue there, as the site itself does.
  const { SITE_URL } = loadEnv(mode, import.meta.dirname, "");

  return {
    server: {
      port: 4200,
      proxy: SITE_URL ? { "/catalog": { target: SITE_URL, changeOrigin: true } } : {},
    },
    resolve: {
      alias: {
        "~": resolve(import.meta.dirname, "./src"),
      },
    },
    plugins: [tanstackRouter({ target: "react", autoCodeSplitting: true }), viteReact(), tailwindcss()],
  };
});
