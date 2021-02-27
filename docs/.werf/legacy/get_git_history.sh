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
#  JSON array will have elements like this:
#
#   {
#     "ts": "1576153516",
#     "date": "2019-12-12T12:25:16Z",
#     "group": "1.0",
#     "channels": [
#       {
#         "name": "alpha",
#         "version": "v1.0.6-rc.6"
#        },
#       {
#         "name": "beta",
#         "version": "v1.0.6-rc.6"
#       },
#       {
#         "name": "rc",
#          "version": "v1.0.6-rc.6"
#       },
#       {
#         "name": "ea",
#         "version": "v1.0.6-rc.6"
#       },
#       {
#          "name": "stable",
#         "version": "v1.0.6-rc.6"
#       }
#     ]
#   }
#
############################################

_PWD=$PWD

WORKDIR=$(mktemp -d -p /tmp/)
REPO=${1:-https://github.com/werf/werf.git}
BRANCH=${2:-multiwerf}

git clone -q -b $BRANCH --single-branch $REPO $WORKDIR
test $? -gt 0  && exit 1

cd $WORKDIR

test $? -gt 0  && exit 1
_OUT=''

for i in $(git log --format="%H-%at" multiwerf.json); do
  COMMIT_HASH=$( echo $i | cut -d- -f1 )
  COMMIT_AUTH_TS=$( echo $i | cut -d- -f2 )
  git show $COMMIT_HASH:multiwerf.json | jq -cM ".multiwerf[] | select( (.outdated != "true") and ( .group | test(\"^1.0\") | not ) ) | {\"ts\":\"$COMMIT_AUTH_TS\",\"date\":\"\($COMMIT_AUTH_TS | tonumber| todate)\",\"group\":.group,\"channels\":[(.channels[] | select(.version != \"v1.2.0-alpha1\") | select(.version != \"v1.2.0-alpha2\"))]} "

done

if [ -n $WORKDIR ]; then  rm -rf $WORKDIR; fi

cd $_PWD
