.PHONY: all werf buildah-test clean

all: werf

werf:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/werf

buildah-test:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/buildah-test


fmt:
	gci -w -local github.com/werf/ pkg/ cmd/ test/
	gofumpt -w cmd/ pkg/

lint:
	golangci-lint -E bidichk -E errname run ./... --build-tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build"


clean:
	rm /home/distorhead/go/bin/werf
	rm /home/distorhead/go/bin/buildah-test
