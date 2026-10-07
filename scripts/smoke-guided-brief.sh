#!/usr/bin/env bash
set -euo pipefail

scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT

set +e
(
  sleep 1
  printf '\r\r\r\r'
  sleep 1
  printf 'guided draft'
  sleep 1
  printf '\003'
) | timeout 12 script -qfec 'stty cols 110 rows 36; go run ./cmd/guided-brief' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [[ $exit_status -ne 0 ]]; then
  cat "$scratch_dir/console"
  exit "$exit_status"
fi

if ! rg -q 'GUIDED BRIEF|What are you writing' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Guided Brief smoke test did not render its form or editor.' >&2
  exit 1
fi

echo 'Guided Brief terminal smoke test passed.'
