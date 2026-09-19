#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
spec="${repo_root}/openapi/openapi.yaml"
bundle="$(mktemp "${TMPDIR:-/tmp}/zvideo-openapi.XXXXXX.json")"
trap 'rm -f "$bundle"' EXIT

redocly lint "$spec"
redocly bundle "$spec" --dereferenced --ext json --output "$bundle"
node "${repo_root}/scripts/validate_openapi_examples.mjs" "$bundle"
