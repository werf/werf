#!/bin/sh
set -eu
case "$1" in
  images)
    test "$2" = --filter
    test "$3" = label=werf=werf-test-one
    if [ -f "$0.removed-id" ]; then echo final; else echo list; fi >> "$0.listings"
    cat "$CLEANUP_INVENTORY"
    ;;
  image)
    test "$2" = inspect
    shift 2
    printf '%s\n' "$@" > "$0.ids"
    if [ -f "$0.inspect" ]; then cat "$0.inspect"; else echo '[]'; fi
    if [ -s "$0.error" ]; then cat "$0.error" >&2; fi
    exit "${CLEANUP_INSPECT_EXIT:-0}"
    ;;
  rmi)
    shift
    test "${1:-}" = --no-prune
    shift
    printf '%s\n' "$1" >> "$CLEANUP_CALLS"
    if [ "${CLEANUP_FAIL:-}" = 1 ]; then exit 42; fi
    if [ "${CLEANUP_NOOP:-}" != 1 ]; then
      printf '%s' "$CLEANUP_REMAINING" > "$CLEANUP_INVENTORY"
    fi
    ;;
  *) exit 43 ;;
esac
