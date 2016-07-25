GITARTIFACT_VERSION = $(shell cat config/projects/dapp-gitartifact.rb | \
                              grep build_version | \
                              grep -oP '[0-9.]+')

DOCKER_IMAGE_VERSION=$(GITARTIFACT_VERSION)
DOCKER_IMAGE_NAME=dappdeps/gitartifact:$(DOCKER_IMAGE_VERSION)

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

all: $(HUB_IMAGE_FILE_PATH)

build/gitartifact_$(GITARTIFACT_VERSION).deb:
	@rm -f pkg/dapp-gitartifact_$(GITARTIFACT_VERSION)*.deb
	@omnibus build -o append_timestamp:false dapp-gitartifact
	@cp pkg/dapp-gitartifact_$(GITARTIFACT_VERSION)-1_amd64.deb \
      build/gitartifact_$(GITARTIFACT_VERSION).deb

build/gitartifact_$(GITARTIFACT_VERSION): build/gitartifact_$(GITARTIFACT_VERSION).deb
	dpkg -x build/gitartifact_$(GITARTIFACT_VERSION).deb build/gitartifact_$(GITARTIFACT_VERSION)

build/Dockerfile_$(GITARTIFACT_VERSION): build/gitartifact_$(GITARTIFACT_VERSION)
	@echo "FROM scratch" > build/Dockerfile_$(GITARTIFACT_VERSION)
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile_$(GITARTIFACT_VERSION)
	@echo "ADD gitartifact_$(GITARTIFACT_VERSION) /" >> build/Dockerfile_$(GITARTIFACT_VERSION)

$(IMAGE_FILE_PATH): build/Dockerfile_$(GITARTIFACT_VERSION)
	docker build -t $(DOCKER_IMAGE_NAME) -f build/Dockerfile_$(GITARTIFACT_VERSION) build
	@echo $(DOCKER_IMAGE_NAME) > $(IMAGE_FILE_PATH)

$(HUB_IMAGE_FILE_PATH): $(IMAGE_FILE_PATH)
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > $(HUB_IMAGE_FILE_PATH)

clean:
	@rm -rf build pkg

.PHONY: all clean omnibus
