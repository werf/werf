ENV_VERSION=0.1.0
DOCKER_IMAGE_NAME=dapp2/env:$(ENV_VERSION)

all: build/hub_image

build:
	@mkdir -p build

build/git: build
	@mkdir -p build/git/aaa
	@echo build/git #TODO

build/sudo: build
	@mkdir -p build/sudo/aaa
	@echo build/sudo #TODO

build/Dockerfile: build/git build/sudo
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	#TODO

build/image: build/Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME) build
	@echo $(DOCKER_IMAGE_NAME) > build/image

build/hub_image: build/image
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > build/hub_image

clean:
	@rm -rf build

.PHONY: all clean
