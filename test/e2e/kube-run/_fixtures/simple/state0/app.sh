#!/bin/bash

set -eu

is_terminated=0

log () {
  echo "[$(date +"%T")] $1"
}

signal_handler () {
  log "Signal handled"
  is_terminated=1
}

trap signal_handler SIGINT SIGTERM EXIT

while [ ${is_terminated} -eq 0 ]; do
  log "Looping ..."
  sleep 1
done

log "Script completed"
