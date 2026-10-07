#!/usr/bin/env bash
set -euo pipefail

scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT

set +e
(
  sleep 1
  printf '\010'
  sleep 1
  printf '\003'
) | timeout 10 script -qfec 'stty cols 140 rows 40; go run ./cmd/ops-deck' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [[ $exit_status -ne 0 ]]; then
  cat "$scratch_dir/console"
  exit "$exit_status"
fi

if ! rg -q 'OPS DECK' "$scratch_dir/transcript" || ! rg -q 'OUTLINE' "$scratch_dir/transcript" || ! rg -q 'LIVE DRAFT' "$scratch_dir/transcript"; then
  cat "$scratch_dir/console"
  echo 'Ops Deck smoke test did not render its dense layout.' >&2
  exit 1
fi

echo 'Ops Deck terminal smoke test passed.'
