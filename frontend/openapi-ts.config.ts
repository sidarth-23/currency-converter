import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../contract/currency-watcher/openapi.yaml",
  output: "src/lib/api/generated",
  plugins: ["@hey-api/client-fetch", "@tanstack/react-query"],
})
