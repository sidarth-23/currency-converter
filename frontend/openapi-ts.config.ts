import { defineConfig } from "@hey-api/openapi-ts"

const apiBaseURL = (
  process.env.VITE_API_BASE_URL ?? "http://localhost:8080"
).replace(/\/$/, "")

export default defineConfig({
  input: `${apiBaseURL}/api/openapi.json`,
  output: "src/lib/api/generated",
  plugins: ["@hey-api/client-fetch", "@tanstack/react-query"],
})
