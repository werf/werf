#!/usr/bin/env bash
set -e

#BINTRAY_AUTH=            # bintray auth user:TOKEN
BINTRAY_SUBJECT=flant     # bintray organization
BINTRAY_REPO=werf         # bintray repository
BINTRAY_PACKAGE=werf      # bintray package in repository

#NO_PRERELEASE=           # This is not a pre release

GITHUB_OWNER=flant     # github user/org
GITHUB_REPO=werf       # github repository

RELEASE_BUILD_DIR=$(pwd)/release/build

GIT_REMOTE=origin      # can be changed to upstream with env

# Werf publisher utility
# Create github release and upload go binary as asset.
main() {
  parse_args "$@" || (usage && exit 1)

  if [ -z "$BINTRAY_AUTH" -a -z "$GITHUB_TOKEN" ] ; then
    echo "Warning! No tokens specified!"
    echo
  fi

  # get git path
  gitPath=
  check_git || (echo "$0: cannot find git command" && exit 2)

  curlPath=
  check_curl || (echo "$0: cannot find curl command" && exit 2)

  TAG_LOCAL_SHA=$($gitPath for-each-ref --format='%(objectname)' refs/tags/$GIT_TAG)
  TAG_REMOTE_SHA=$($gitPath ls-remote --tags $GIT_REMOTE refs/tags/$GIT_TAG | cut -f 1)

  if [ "x$TAG_LOCAL_SHA" != "x$TAG_REMOTE_SHA" ] ; then
    echo "CRITICAL: Tag $GIT_TAG should be pushed to $GIT_REMOTE before creating new release"
    exit 1
  fi

  $gitPath checkout -f $GIT_TAG || (echo "$0: git checkout error" && exit 2)

  # version for release without v prefix
  VERSION=${GIT_TAG#v}
  # message for github release and bintray version description
  # change to *contents to get commit message
  TAG_RELEASE_MESSAGE=$($gitPath for-each-ref --format="%(contents)" refs/tags/$GIT_TAG | jq -R -s '.' )

  source $(dirname $0)/build.sh $GIT_TAG

  echo "Publish version $VERSION from git tag $GIT_TAG"
  if [ -n "$BINTRAY_AUTH" ] ; then
    ( bintray_create_version && echo "Bintray: Version $VERSION created" ) || ( exit 1 )

    for filename in werf werf.sha ; do
      for arch in darwin linux ; do
        ( bintray_upload_file_into_version $RELEASE_BUILD_DIR/$arch-amd64/$filename $arch-amd64/$filename ) || ( exit 1 )
      done
    done
  fi

  if [ -n "$GITHUB_TOKEN" ] ; then
    ( github_create_release && echo "Github: Release for tag $GIT_TAG created" ) || ( exit 1 )
  fi

}

bintray_create_version() {
PAYLOAD=$(cat <<- JSON
  {
     "name": "${VERSION}",
     "desc": ${TAG_RELEASE_MESSAGE},
     "vcs_tag": "${GIT_TAG}"
  }
JSON
)
  curlResponse=$(mktemp)
  status=$(curl -s -w %{http_code} -o $curlResponse \
      --request POST \
      --user $BINTRAY_AUTH \
      --header "Content-type: application/json" \
      --data "$PAYLOAD" \
      https://api.bintray.com/packages/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/versions
  )

  echo "Bintray create version: curl return status $status with response"
  cat $curlResponse
  echo
  rm $curlResponse

  ret=0
  if [ "x$(echo $status | cut -c1)" != "x2" ]
  then
    ret=1
  fi

  return $ret
}

# upload file to $GIT_TAG version
bintray_upload_file_into_version() {
  UPLOAD_FILE_PATH=$1
  DESTINATION_PATH=$2

  curlResponse=$(mktemp)
  status=$(curl -s -w %{http_code} -o $curlResponse \
      --header "X-Bintray-Publish: 1" \
      --header "Content-type: application/binary" \
      --request PUT \
      --user $BINTRAY_AUTH \
      --upload-file $UPLOAD_FILE_PATH \
      https://api.bintray.com/content/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/$VERSION/$VERSION/$DESTINATION_PATH
  )

  echo "Bintray upload $DESTINATION_PATH: curl return status $status with response"
  cat $curlResponse
  echo
  rm $curlResponse

  ret=0
  if [ "x$(echo $status | cut -c1)" != "x2" ]
  then
    ret=1
  else
    dlUrl="https://dl.bintray.com/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${VERSION}/${DESTINATION_PATH}"
    echo "Bintray: $DESTINATION_PATH uploaded to ${dlURL}"
  fi

  return $ret
}

github_create_release() {
  prerelease="true"
  if [[ "$NO_PRERELEASE" == "yes" ]] ; then
    prerelease="false"
    echo "# Creating release $GIT_TAG"
  else
    echo "# Creating pre-release $GIT_TAG"
  fi

  GHPAYLOAD=$(cat <<- JSON
{
  "tag_name": "$GIT_TAG",
  "name": "Werf $VERSION",
  "body": $TAG_RELEASE_MESSAGE,
  "draft": false,
  "prerelease": $prerelease
}
JSON
)

  curlResponse=$(mktemp)
  status=$(curl -s -w %{http_code} -o $curlResponse \
      --request POST \
      --header "Authorization: token $GITHUB_TOKEN" \
      --header "Accept: application/vnd.github.v3+json" \
      --data "$GHPAYLOAD" \
      https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases
  )

  echo "Github create release: curl return status $status with response"
  cat $curlResponse
  echo
  rm $curlResponse

  ret=0
  if [ "x$(echo $status | cut -c1)" != "x2" ]
  then
    ret=1
  fi

  return $ret
}

check_git() {
  gitPath=$(which git) || return 1
}

check_curl() {
  curlPath=$(which curl) || return 1
}



usage() {
printf " Usage: $0 --tag <tagname> [--github-token TOKEN] [--bintray-token TOKEN]

    --no-prerelease
            This is final release, not a prerelease. Prerelease will be created by default.
            Can be changed manually later in the github UI.

    --tag
            Release is a tag based. Tag should be present if gh-token specified.

    --github-token TOKEN
            Write access token for github. No github actions if no token specified.

    --bintray-auth user:TOKEN
            User and token for upload to bintray.com. No bintray actions if no token specified.

    --help|-h
            Print help

"
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --tag)
        GIT_TAG="$2"
				shift
        ;;
      --github-token)
        GITHUB_TOKEN="$2"
        shift
        ;;
      --bintray-auth)
        BINTRAY_AUTH="$2"
        shift
        ;;
      --no-prerelease)
        NO_PRERELEASE="yes"
        ;;
      --help|-h)
        return 1
        ;;
      --*)
        echo "Illegal option $1"
        return 1
        ;;
    esac
    shift $(( $# > 0 ? 1 : 0 ))
  done

  [ -z "$GIT_TAG" ] && return 1 || return 0
}

# wait for full file download if executed as
# $ curl | sh
main "$@"
