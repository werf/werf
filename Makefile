ENV_VERSION=0.1.0
DOCKER_IMAGE_NAME=dapp2/env:$(ENV_VERSION)

all: omnibus

omnibus:
	@omnibus build dapp-env

clean:
	@rm -rf build pkg

.PHONY: all clean

#FIXME >>>
build/git:
	@mkdir -p build/git
	@echo build/git #TODO

build/sudo:
	@mkdir -p build/sudo
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
