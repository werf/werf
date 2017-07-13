#!/bin/bash

set -e

while getopts "v:" opt; do
  case "$opt" in
    v)
      DAPP_VERSION=$OPTARG
      ;;
  esac
done

gem install dapp $(if [[ $DAPP_VERSION ]] ; then echo "--version=$DAPP_VERSION" ; fi)

if [ -d "/etc/cron.d" ] ; then
  echo
  echo "Installing dapp update cron job"
  echo "* * * * * $USER /bin/bash -lec 'dapp update'" | sudo tee /etc/cron.d/dapp-update
  echo
fi

