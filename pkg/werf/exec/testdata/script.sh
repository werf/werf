#!/bin/bash

set -eu

termination_enabled=0

script_timeout=$1
script_exit_code=$2

function sigterm_handler () {
  echo "[script]: sigterm handled"
  termination_enabled=1
}

function sigint_handler () {
  echo "[script]: sigint handled"
  termination_enabled=1
}

trap sigterm_handler SIGTERM
trap sigint_handler SIGINT

while [ ${termination_enabled} -eq 0 ]; do
  echo "[script]: Looping infinitely..."
  sleep 1s
done

sleep "${script_timeout}"
echo "[script]: Completed."
exit "${script_exit_code}"
