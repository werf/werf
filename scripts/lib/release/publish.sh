publish_binaries() {
    VERSION=$1

    echo "Publish version $VERSION from git tag $VERSION"
    ( bintray_create_version $VERSION && echo "Bintray: Version $VERSION created" ) || ( exit 1 )

    for os in linux darwin windows ; do
        for arch in amd64 ; do
          fileName=werf-$os-$arch-$VERSION
          if [ "$os" == "windows" ] ; then
              fileName=$fileName.exe
          fi

          localFile=$RELEASE_BUILD_DIR/$VERSION/$fileName

          ( bintray_upload_file_into_version $VERSION $localFile $fileName ) || ( exit 1 )
        done
      done

    ( bintray_upload_file_into_version $VERSION $RELEASE_BUILD_DIR/$VERSION/SHA256SUMS SHA256SUMS ) || ( exit 1 )
}

sign_binaries() {
    VERSION=$1

    dlUrl="https://dl.bintray.com/${BINTRAY_SUBJECT}/${BINTRAY_REPO}/${VERSION}/SHA256SUMS"

    signDir="$RELEASE_BUILD_DIR/$VERSION/sign"

    mkdir -p $signDir

    curl $dlUrl -o $signDir/SHA256SUMS

    # TODO
    cp $signDir/SHA256SUMS $signDir/SHA256SUMS.dsc

    ( bintray_upload_file_into_version $VERSION $signDir/SHA256SUMS.dsc SHA256SUMS.dsc ) || ( exit 1 )
}
