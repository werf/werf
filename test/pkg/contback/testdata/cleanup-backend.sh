#!/bin/sh
set -eu
case "$1" in
  images)
    test "$2" = --filter
    test "$3" = label=werf=werf-test-one
    cat "$CLEANUP_INVENTORY"
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
