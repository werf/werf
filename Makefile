ENV_VERSION=0.1.0
DOCKER_IMAGE_NAME=dapp2/env:$(ENV_VERSION)

PKG_BUILD_VERSION = $(shell cat pkg/version-manifest.json | \
														grep -oP 'build_version":"[0-9.+]+"' | \
														cut -d'"' -f3)
PKG_BUILD_ITERATION = $(shell cat config/projects/dapp-env.rb  | \
															grep build_iteration | \
															cut -d' ' -f2)
ENV_DEB_PATH = $(shell ls -1 pkg/dapp-env_$(PKG_BUILD_VERSION)-$(PKG_BUILD_ITERATION)_*.deb | \
											 tail -n1)

all: build/hub_image

pkg/version-manifest.json:
	@omnibus build dapp-env

build/dapp-env: pkg/version-manifest.json
	dpkg -x $(ENV_DEB_PATH) build/dapp-env

build/Dockerfile: build/dapp-env
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	@echo "ADD dapp-env /" >> build/Dockerfile

build/image: build/Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME) build
	@echo $(DOCKER_IMAGE_NAME) > build/image

build/hub_image: build/image
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > build/hub_image

clean:
	@rm -rf build pkg

.PHONY: all clean
