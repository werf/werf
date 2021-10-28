#!/bin/bash

set -e

############################################
##
## Prints JSON array from git history
##
############################################
#
#  Syntax: $0 [git-repo] [git-branch]
#
############################################

_PWD=$PWD

WORKDIR=$(mktemp -d -p /tmp/)
REPO=${1:-https://github.com/werf/werf.git}
BRANCH=${2:-multiwerf}

if [[ -x /usr/local/bin/yq ]]; then
  YQ=/usr/local/bin/yq
else
  type yq &>/dev/null
  test $? -gt 0  && exit 1
  YQ=yq
fi

git clone -q -b $BRANCH --single-branch $REPO $WORKDIR
test $? -gt 0  && exit 1

cd $WORKDIR

test $? -gt 0  && exit 1
_OUT=''

for i in $(git log --format="%H-%at" -- trdl_channels.yaml multiwerf.json ); do
  COMMIT_HASH=$( echo $i | cut -d- -f1 )
  COMMIT_AUTH_TS=$( echo $i | cut -d- -f2 )
  CONTENT=$(2>/dev/null git show $COMMIT_HASH:trdl_channels.yaml | $YQ eval -o json)
  if [[ "$CONTENT" != "null" ]]; then
      CONTENT=$(echo $CONTENT | jq '.groups[] | {"group":.name, "channels": .channels}' | jq -s ' {"multiwerf":.}')
  else
      CONTENT=$(git show $COMMIT_HASH:multiwerf.json)
  fi
  echo $CONTENT | jq -cM ".multiwerf[] | select( (.outdated != "true") and ( .group | test(\"^1.0\") | not ) ) | {\"ts\":\"$COMMIT_AUTH_TS\",\"date\":\"\($COMMIT_AUTH_TS | tonumber| todate)\",\"group\":.group,\"channels\":[(.channels[] | select(.version != \"v1.2.0-alpha1\") | select(.version != \"v1.2.0-alpha2\"))]} "
done

if [ -n $WORKDIR ]; then  rm -rf $WORKDIR; fi

cd $_PWD
