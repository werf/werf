.PHONY: all werf buildah-test fmt lint docs clean build-images

ifeq ($(OS),Windows_NT)
    uname_S := Windows
else
    uname_S := $(shell uname -s)
endif

all: werf

ifeq ($(uname_S), Linux)
# Cgo needed for proper buildah support
werf: werf-with-cgo
else
werf: werf-without-cgo
endif

werf-with-cgo:
	CGO_ENABLED=1 go install -compiler gc -ldflags="-linkmode external -extldflags=-static" -tags="dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build" github.com/werf/werf/cmd/werf

werf-without-cgo:
	CGO_ENABLED=0 go install -tags="dfrunmount dfssh containers_image_openpgp" "github.com/werf/werf/cmd/werf"

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

docs: werf
	./scripts/docs/regen.sh $$GOPATH/bin/werf

docs_check_broken_links_ru: werf-without-cgo
	./scripts/docs/check_broken_links.sh ru $$GOPATH/bin/werf

docs_check_broken_links_en: werf-without-cgo
	./scripts/docs/check_broken_links.sh main $$GOPATH/bin/werf

build-images:
	cd ./scripts/images && PACKER_LOG=1 packer build -force template.pkr.hcl

clean:
	rm -f $$GOPATH/bin/werf
	rm -f $$GOPATH/buildah-test
