#!/bin/sh
set -e

SCRIPT="$(dirname "$0")/../install.sh"

output=$(VERSION=1.0.0 sh "$SCRIPT" --dry-run 2>&1)

echo "$output" | grep -q "OS:" || { echo "FAIL: OS not detected"; exit 1; }
echo "$output" | grep -q "Arch:" || { echo "FAIL: Arch not detected"; exit 1; }
echo "$output" | grep -q "logsage_1.0.0" || { echo "FAIL: artifact name not in output"; exit 1; }

echo "PASS: install.sh dry-run smoke test"
