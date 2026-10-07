#!/usr/bin/env bash
set -euo pipefail

scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT

set +e
(
  sleep 1
  printf '\013'
  sleep 1
  printf '\003'
) | timeout 10 script -qfec 'stty cols 110 rows 36; go run ./cmd/color-field' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [[ $exit_status -ne 0 ]]; then
  cat "$scratch_dir/console"
  exit "$exit_status"
fi

if ! rg -q 'COLOR FIELD' "$scratch_dir/transcript" || ! rg -q 'COMMAND PALETTE' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Color Field smoke test did not render the editor and palette.' >&2
  exit 1
fi

if ! rg -q $'\033\\[38;5;' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Color Field smoke test did not emit ANSI-256 color sequences.' >&2
  exit 1
fi

echo 'Color Field terminal smoke test passed.'
