bintray_create_version() {
    VERSION=$1

    PAYLOAD=$(cat <<- JSON
    {
        "name": "${VERSION}",
        "desc": "${VERSION}",
        "vcs_tag": "${GIT_TAG}"
    }
JSON
)

    curlResponse=$(mktemp)
    status=$(curl -s -w %{http_code} -o $curlResponse \
        --request POST \
        --user $PUBLISH_BINTRAY_AUTH \
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
    VERSION=$1
    UPLOAD_FILE_PATH=$2
    DESTINATION_PATH=$3

    curlResponse=$(mktemp)
    status=$(curl -s -w %{http_code} -o $curlResponse \
        --header "Content-type: application/binary" \
        --request PUT \
        --user $PUBLISH_BINTRAY_AUTH \
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

bintray_publish_files_in_version() {
    local VERSION=$1

    curlResponse=$(mktemp)
    status=$(curl -s -w '%{http_code}' -o "$curlResponse" \
        --request POST \
        --user "$PUBLISH_BINTRAY_AUTH" \
        --header "Content-type: application/json" \
        "https://api.bintray.com/content/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${VERSION}/publish"
    )

    echo "Bintray publish files in version ${VERSION}: curl return status $status with response"
    cat "$curlResponse"
    echo
    rm "$curlResponse"

    ret=0
    if [ "x$(echo "$status" | cut -c1)" != "x2" ]
    then
      ret=1
    fi

    return $ret
}

