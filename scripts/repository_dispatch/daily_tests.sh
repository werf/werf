#!/bin/bash -e

if [ -z "$GITHUB_TOKEN" ] ; then
    echo "Required GITHUB_TOKEN (https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line#creating-a-token)!" 1>&2
    exit 1
fi

curl \
  --location --request POST 'https://api.github.com/repos/werf/werf/dispatches' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/vnd.github.everest-preview+json' \
  --header "Authorization: token $GITHUB_TOKEN" \
  --data-raw '{
    "event_type": "daily_tests",
    "client_payload": {}
  }'
