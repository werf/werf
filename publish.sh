#!/usr/bin/env bash
set -e

#BINTRAY_AUTH=          # bintray auth user:TOKEN
BINTRAY_SUBJECT=flant  # bintray organization
BINTRAY_REPO=dapp      # bintray repository
BINTRAY_PACKAGE=ruby2go  # bintray package in repository

GITHUB_OWNER=flant     # github user/org
GITHUB_REPO=dapp       # github repository

RUBY2GO_BINARIES_NAMES=$(ls cmd)
UPLOAD_FROM_DIR=$GOPATH/bin

GIT_REMOTE=origin      # can be changed to upstream with env

# Dapp publisher utility
# Publish gem to RubyGems. Create github release and upload go binary as asset.
main() {
  parse_args "$@" || (usage && exit 1)

  if [ -z "$BINTRAY_AUTH" -a -z "$GITHUB_TOKEN" -a -z "$RUBYGEMS_TOKEN" ] ; then
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

  build_gem && echo "Build gem is successful" || ( exit 1 )
  calculate_binaries_checksums && echo "Binaries checksums calculation is successful" || ( exit 1 )

  echo "Publish version $VERSION from git tag $GIT_TAG"
  if [ -n "$BINTRAY_AUTH" ] ; then
    ( bintray_create_version && echo "Bintray: Version $VERSION created" ) || ( exit 1 )

    for bin in $RUBY2GO_BINARIES_NAMES ; do
      ( bintray_upload_file $bin ) || ( exit 1 )
      ( bintray_upload_file $bin.sha ) || ( exit 1 )
    done
  fi

  if [ -n "$GITHUB_TOKEN" ] ; then
    ( github_create_release && echo "Github: Release for tag $GIT_TAG created" ) || ( exit 1 )
  fi

  if [ -n "$RUBYGEMS_TOKEN" ] ; then
    ( rubygems_upload_gem && echo "Rubygems: Gem uploaded" ) || ( exit 1 )
  fi

}

build_gem() {
  echo "Building gem dapp"
  rm -f dapp-*.gem
  gem build ./dapp.gemspec

  GEM_FILE_PATH=$(ls -1 dapp-*.gem | head -n 1)
}

calculate_binaries_checksums() {
  echo "Calculating ruby2go binaries checksums"

  for bin in $RUBY2GO_BINARIES_NAMES ; do
    sha256sum $UPLOAD_FROM_DIR/$bin | cut -d' ' -f 1 > $UPLOAD_FROM_DIR/$bin.sha
  done
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
bintray_upload_file() {
  UPLOAD_FILENAME=$1
  curlResponse=$(mktemp)
  status=$(curl -s -w %{http_code} -o $curlResponse \
      --header "X-Bintray-Publish: 1" \
      --header "Content-type: application/binary" \
      --request PUT \
      --user $BINTRAY_AUTH \
      --upload-file $UPLOAD_FROM_DIR/$UPLOAD_FILENAME \
      https://api.bintray.com/content/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/$VERSION/$VERSION/$UPLOAD_FILENAME
  )

  echo "Bintray upload $UPLOAD_FILENAME: curl return status $status with response"
  cat $curlResponse
  echo
  rm $curlResponse

  ret=0
  if [ "x$(echo $status | cut -c1)" != "x2" ]
  then
    ret=1
  else
    dlUrl="https://dl.bintray.com/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${VERSION}/${UPLOAD_FILENAME}"
    echo "Bintray: $UPLOAD_FILENAME uploaded to ${dlURL}"
  fi

  return $ret
}

github_create_release() {
  GHPAYLOAD=$(cat <<- JSON
{
  "tag_name": "$GIT_TAG",
  "name": "$GITHUB_REPO $VERSION",
  "body": $TAG_RELEASE_MESSAGE,
  "draft": false,
  "prerelease": false
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

rubygems_upload_gem() {
  curlResponse=$(mktemp)
  status=$(curl -s -w %{http_code} -o $curlResponse \
      --request POST \
      --header "Authorization: ${RUBYGEMS_TOKEN}" \
      --header "Content-type: application/octet-stream" \
      --data-binary @${GEM_FILE_PATH} \
      https://rubygems.org/api/v1/gems
  )

  echo "Rubygems upload ${GEM_FILE_PATH}: curl return status $status with response"
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
printf " Usage: $0 --tag <tagname> [--github-token TOKEN] [--rubygems-token TOKEN]
                           [--bintray-token TOKEN]

    --tag
            Release is a tag based. Tag should be present if gh-token specified.

    --github-token TOKEN
            Write access token for github. No github actions if no token specified.

    --bintray-auth user:TOKEN
            User and token for upload to bintray.com. No bintray actions if no token specified.

    --rubygems-token TOKEN
            Token for rubygems.org. No upload to rubygems if no token specified

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
      --rubygems-token)
        RUBYGEMS_TOKEN="$2"
        shift
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
