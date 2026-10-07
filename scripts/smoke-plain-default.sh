#!/usr/bin/env bash
# Exercise Plain Default in a real pseudo-terminal, not just its Go model.
set -euo pipefail

scratch_dir=$(mktemp -d)
trap 'rm -rf "$scratch_dir"' EXIT

set +e
(
  sleep 4
  printf 'Hello, terminal editor!'
  sleep 1
  printf '\033[21~'
  sleep 1
  printf '\033[B\r'
  sleep 1
  printf '\003'
) | timeout 15 script -qfec 'stty cols 110 rows 36; go run ./cmd/plain-default' "$scratch_dir/transcript" >"$scratch_dir/console" 2>&1
exit_status=${PIPESTATUS[1]}
set -e

if [ "$exit_status" -ne 0 ]; then
  echo 'editor did not exit cleanly after Ctrl+C' >&2
  exit 1
fi

if rg -q '\]11' "$scratch_dir/transcript"; then
  echo 'background-color probe leaked into editor input' >&2
  exit 1
fi

rg -q 'Hello, terminal editor!' "$scratch_dir/transcript"
rg -q 'File  Edit  Format  Insert  View  Help' "$scratch_dir/transcript"
rg -q 'Open path:' "$scratch_dir/transcript"

echo 'Plain Default terminal smoke test passed.'
