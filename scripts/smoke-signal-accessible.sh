#!/usr/bin/env bash
set -euo pipefail

scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT

set +e
(
  sleep 1
  printf '\005'
  sleep 1
  printf '\003'
) | timeout 10 script -qfec 'stty cols 110 rows 36; go run ./cmd/signal-accessible' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [[ $exit_status -ne 0 ]]; then cat "$scratch_dir/console"; exit "$exit_status"; fi
if ! rg -q 'SIGNAL' "$scratch_dir/transcript" || ! rg -q '\[EDITING\]' "$scratch_dir/transcript" || ! rg -q '\[READING\]' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Signal smoke test did not render explicit accessibility labels.' >&2
  exit 1
fi
echo 'Signal terminal smoke test passed.'
