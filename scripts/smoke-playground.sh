#!/usr/bin/env bash
set -euo pipefail
scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT
set +e
(
  sleep 1
  printf '\030'
  sleep 1
  printf '\003'
) | timeout 10 script -qfec 'stty cols 130 rows 40; go run ./cmd/playground' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e
if [[ $exit_status -ne 0 ]]; then cat "$scratch_dir/console"; exit "$exit_status"; fi
if ! rg -q 'YOUR DRAFT' "$scratch_dir/console" || ! rg -q 'HOW IT READS' "$scratch_dir/console"; then cat "$scratch_dir/console"; echo 'Playground smoke test did not render its expressive layout.' >&2; exit 1; fi
echo 'Playground terminal smoke test passed.'
