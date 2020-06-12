export GO111MODULE=on
export CGO_ENABLED=0

go_mod_download() {
    VERSION=$1

    for os in linux darwin windows ; do
        for arch in amd64 ; do
            echo "# Downloading go modules for GOOS=$os GOARCH=$arch"

            n=0
            until [ $n -gt 5 ]
            do
                ( GOOS=$os GOARCH=$arch go mod download ) && break || true
                n=$[$n+1]

                if [ ! $n -gt 5 ] ; then
                    echo "[$n] Retrying modules downloading"
                fi
            done

            if [ $n -gt 5 ] ; then
                echo "Exiting due to 'go mod download' failures"
                exit 1
            fi
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
              go build -tags "dfrunmount dfssh" -ldflags="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION" \
                       -o $outputFile github.com/werf/werf/cmd/werf

            echo "# Built $outputFile"
        done
    done

    cd $RELEASE_BUILD_DIR/$VERSION/ && \
      sha256sum werf-* > SHA256SUMS && \
      cd -
}
