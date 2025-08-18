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

sanitize() {
  local input="$1"
  local sanitized
  sanitized=$(printf '%s' "$input" | LC_ALL=C tr -cd '[:alnum:]._ +-' )
  if [[ "$input" != "$sanitized" ]]; then
    die "Invalid input detected: $input"
  fi
}

validate_environment() {
  local env="$1"
  case "$env" in
    dev|staging|prod) ;;
    *) die "Invalid environment: $env" ;;
  esac
}

validate_version() {
  local version="$1"
  if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
    die "Invalid version: $version"
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
sanitize "${TEST_FLAGS-}"

# Optional inputs: validate if provided
if [[ -n "${ENVIRONMENT-}" ]]; then
  check_no_ctrl ENVIRONMENT
  sanitize "${ENVIRONMENT}"
  validate_environment "${ENVIRONMENT}"
fi

if [[ -n "${VERSION-}" ]]; then
  check_no_ctrl VERSION
  sanitize "${VERSION}"
  validate_version "${VERSION}"
fi

echo "Inputs validated successfully."
