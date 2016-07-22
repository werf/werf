GITARTIFACT_VERSION = $(shell cat config/projects/dapp-gitartifact.rb | \
                              grep build_version | \
                              grep -oP '[0-9.]+')

DOCKER_IMAGE_VERSION=$(GITARTIFACT_VERSION)
DOCKER_IMAGE_NAME=dappdeps/gitartifact:$(DOCKER_IMAGE_VERSION)

FETCH_PKG_BUILD_VERSION = cat pkg/version-manifest.json | \
                          grep -oP 'build_version":"[0-9.+]+"' | \
                          cut -d'"' -f3 | \
                          grep -P '^$(DOCKER_IMAGE_VERSION)'
FETCH_PKG_BUILD_ITERATION = cat config/projects/dapp-gitartifact.rb  | \
                            grep build_iteration | \
                            cut -d' ' -f2
FETCH_GITARTIFACT_OMNIBUS_DEB_PATH = ls -1 pkg/dapp-gitartifact_`$(FETCH_PKG_BUILD_VERSION)`-`$(FETCH_PKG_BUILD_ITERATION)`_*.deb | \
																		 tail -n1

GITARTIFACT_DEB_PATH=build/gitartifact_$(DOCKER_IMAGE_VERSION).deb

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

all: $(HUB_IMAGE_FILE_PATH)

omnibus:
	@bash -ec 'if [ ! -f $(GITARTIFACT_DEB_PATH) ] ; then omnibus build dapp-gitartifact ; fi'

$(GITARTIFACT_DEB_PATH): omnibus
	@cp $(shell $(FETCH_GITARTIFACT_OMNIBUS_DEB_PATH)) $(GITARTIFACT_DEB_PATH)

build/dapp-gitartifact: $(GITARTIFACT_DEB_PATH)
	dpkg -x $(GITARTIFACT_DEB_PATH) build/dapp-gitartifact

build/Dockerfile: build/dapp-gitartifact
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	@echo "ADD dapp-gitartifact /" >> build/Dockerfile

$(IMAGE_FILE_PATH): build/Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME) build
	@echo $(DOCKER_IMAGE_NAME) > $(IMAGE_FILE_PATH)

$(HUB_IMAGE_FILE_PATH): $(IMAGE_FILE_PATH)
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > $(HUB_IMAGE_FILE_PATH)

clean:
	@rm -rf build pkg

.PHONY: all clean omnibus
