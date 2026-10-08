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
) | timeout 10 script -qfec 'stty cols 120 rows 36; go run ./cmd/windows-desktop' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [[ $exit_status -ne 0 ]]; then
  cat "$scratch_dir/console"
  exit "$exit_status"
fi

if ! rg -q 'F6 List' "$scratch_dir/transcript" || ! rg -q 'Document' "$scratch_dir/transcript" || ! rg -q 'Preview' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Windows Desktop smoke test did not render editor chrome.' >&2
  exit 1
fi

if ! rg -q $'\033\\[46m' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Windows Desktop smoke test did not emit the teal desktop background.' >&2
  exit 1
fi

echo 'Windows Desktop terminal smoke test passed.'
