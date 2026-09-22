#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
destination="$root/contract/frankfurter/v2/openapi.json"
temporary="$(mktemp "${destination}.XXXXXX")"
trap 'rm -f "$temporary"' EXIT

curl --fail --proto '=https' --tlsv1.2 https://api.frankfurter.dev/v2/openapi.json -o "$temporary"
jq -e '.openapi and .paths["/rates"] and .paths["/rates"].get.operationId == "getRates"' "$temporary" >/dev/null
mv "$temporary" "$destination"
