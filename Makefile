.PHONY: all werf buildah-test fmt lint docs clean

all: werf

werf:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/werf

buildah-test:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/buildah-test

integration-tests:
	export WERF_TEST_K8S_DOCKER_REGISTRY="localhost:5000"
	export WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME="nobody"
	export WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD=""
	CGO_ENABLED=1 ginkgo -r -v -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" integration/suites

e2e-tests:
	export WERF_TEST_K8S_DOCKER_REGISTRY="localhost:5000"
	export WERF_TEST_K8S_DOCKER_REGISTRY_USERNAME="nobody"
	export WERF_TEST_K8S_DOCKER_REGISTRY_PASSWORD=""
	CGO_ENABLED=1 ginkgo -r -v -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" test/e2e

.PHONY: unit-tests
unit-tests:
	CGO_ENABLED=1 go test -v -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/pkg/...

fmt:
	gci -w -local github.com/werf/ pkg/ cmd/ test/
	gofumpt -w cmd/ pkg/ test/ integration/

lint:
	golangci-lint run ./... --build-tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build"


docs:
	./docs/regen.sh


clean:
	rm -f $$GOPATH/bin/werf
	rm -f $$GOPATH/buildah-test
