#!/usr/bin/env bash
set -euo pipefail

# Reusable CI input validator for GitHub Actions
# Fails fast on unexpected or unsafe input values.

die() { echo "Input validation error: $*" >&2; exit 2; }

# Validate that inputs contain no control characters
check_no_ctrl() {
  local name="$1" val
  val="${!name-}"
  if printf '%s' "$val" | LC_ALL=C grep -q '[[:cntrl:]]'; then
    die "$name contains control characters"
  fi
}

# --- Known inputs ---
# TEST_FLAGS: allowed values are empty or "--released"
check_test_flags() {
  local val="${TEST_FLAGS-}"
  case "$val" in
    ""|"--released") ;;
    *) die "TEST_FLAGS must be one of: '' or '--released' (got: '$val')" ;;
  esac
}

# Run checks
check_no_ctrl TEST_FLAGS
check_test_flags

echo "Inputs validated successfully."
