.PHONY: all werf buildah-test unit-test fmt lint clean

all: werf

werf:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/werf

buildah-test:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/buildah-test


unit-test:
	CGO_ENABLED=1 go test -v -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/pkg/...

fmt:
	gci -w -local github.com/werf/ pkg/ cmd/ test/
	gofumpt -w cmd/ pkg/

lint:
	golangci-lint -E bidichk -E errname run ./... --build-tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build"


clean:
	rm -f $$GOPATH/bin/werf
	rm -f $$GOPATH/buildah-test
