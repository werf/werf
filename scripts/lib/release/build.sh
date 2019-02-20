go_get() {
    for os in linux darwin windows ; do
        for arch in amd64 ; do
            export GOOS=$os
            export GOARCH=$arch
            source $GOPATH/src/github.com/flant/werf/go-get.sh
        done
    done
}

go_build() {
    VERSION=$1

    rm -rf $RELEASE_BUILD_DIR/$VERSION
    mkdir -p $RELEASE_BUILD_DIR/$VERSION
    chmod -R 0777 $RELEASE_BUILD_DIR/$VERSION

    for os in linux darwin windows ; do
        for arch in amd64 ; do
            outputFile=$RELEASE_BUILD_DIR/$VERSION/werf-$os-$arch-$VERSION
            if [ "$os" == "windows" ] ; then
                outputFile=$outputFile.exe
            fi

            echo "# Building werf $VERSION for $os $arch ..."

            GOOS=$os GOARCH=$arch \
              go build -ldflags="-s -w -X github.com/flant/werf/pkg/werf.Version=$VERSION" \
                       -o $outputFile github.com/flant/werf/cmd/werf

            echo "# Built $outputFile"
        done
    done

    cd $RELEASE_BUILD_DIR/$VERSION/
    sha256sum werf-* > SHA256SUMS
    cd -
}

build_binaries() {
    VERSION=$1

    # git checkout -f $VERSION || (echo "$0: git checkout error" && exit 2)

    ( go_get ) || ( exit 1 )
    ( go_build $VERSION ) || ( exit 1 )
}
