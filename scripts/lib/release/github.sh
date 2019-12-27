create_github_release() {
    VERSION=$1

    TAG_RELEASE_MESSAGE=$(git for-each-ref --format="%(contents)" refs/tags/$VERSION | jq -R -s '.' )

    GHPAYLOAD=$(cat <<- JSON
    {
    "tag_name": "$VERSION",
    "name": "werf $VERSION",
    "body": $TAG_RELEASE_MESSAGE,
    "draft": false,
    "prerelease": false
    }
JSON
)

    curlResponse=$(mktemp)
    status=$(curl -s -w %{http_code} -o $curlResponse \
        --request POST \
        --header "Authorization: token $PUBLISH_GITHUB_TOKEN" \
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
