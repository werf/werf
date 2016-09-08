DOCKER_IMAGE_VERSION = $(shell cat config/projects/dappdeps-base.rb | \
                               grep build_version | \
                               grep -oP '[0-9.]+')
DOCKER_IMAGE_NAME=dappdeps/base:$(DOCKER_IMAGE_VERSION)

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

all: $(HUB_IMAGE_FILE_PATH)

build/dappdeps-base_$(DOCKER_IMAGE_VERSION).deb:
	@rm -f pkg/dappdeps-base_$(DOCKER_IMAGE_VERSION)*.deb
	@omnibus build -o append_timestamp:false dappdeps-base
	@cp pkg/dappdeps-base_$(DOCKER_IMAGE_VERSION)-1_amd64.deb \
      build/dappdeps-base_$(DOCKER_IMAGE_VERSION).deb

build/dappdeps-base_$(DOCKER_IMAGE_VERSION): build/dappdeps-base_$(DOCKER_IMAGE_VERSION).deb
	dpkg -x build/dappdeps-base_$(DOCKER_IMAGE_VERSION).deb build/dappdeps-base_$(DOCKER_IMAGE_VERSION)

build/Dockerfile_$(DOCKER_IMAGE_VERSION): build/dappdeps-base_$(DOCKER_IMAGE_VERSION)
	@echo "FROM scratch" > build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	@echo "ADD dappdeps-base_$(DOCKER_IMAGE_VERSION) /" >> build/Dockerfile_$(DOCKER_IMAGE_VERSION)

$(IMAGE_FILE_PATH): build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	docker build -t $(DOCKER_IMAGE_NAME) -f build/Dockerfile_$(DOCKER_IMAGE_VERSION) build
	@echo $(DOCKER_IMAGE_NAME) > $(IMAGE_FILE_PATH)

$(HUB_IMAGE_FILE_PATH): $(IMAGE_FILE_PATH)
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > $(HUB_IMAGE_FILE_PATH)

clean:
	@rm -rf build pkg

.PHONY: all clean
