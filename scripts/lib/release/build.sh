go_build() {
	export GO111MODULE="on"

	COMMON_LDFLAGS="-s -w -X github.com/werf/werf/pkg/werf.Version=$VERSION"
	COMMON_TAGS="dfrunmount dfssh containers_image_openpgp"
	PKG="github.com/werf/werf/cmd/werf"

	parallel -j0 --halt now,fail=1 --line-buffer -k --tag --tagstring '{= @cmd = split(" ", $_); $_ = "[".$cmd[0]." ".$cmd[1]."]" =}' <<-EOF
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
    go build -compiler gc -o "release-build/$VERSION/werf-linux-amd64-$VERSION" \
    -ldflags="$COMMON_LDFLAGS -linkmode external -extldflags=-static" \
    -tags="$COMMON_TAGS osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    "$PKG"

  GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
    go build -compiler gc -o "release-build/$VERSION/werf-linux-arm64-$VERSION" \
    -ldflags="$COMMON_LDFLAGS -linkmode external -extldflags=-static" \
    -tags="$COMMON_TAGS osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" \
    "$PKG"

  GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/werf-darwin-amd64-$VERSION" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"

  GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/werf-darwin-arm64-$VERSION" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"

  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "release-build/$VERSION/werf-windows-amd64-$VERSION" -ldflags="$COMMON_LDFLAGS" -tags="$COMMON_TAGS" "$PKG"
EOF

	cd release-build/$VERSION/ && \
		sha256sum werf-* > SHA256SUMS && \
		cd -
}
