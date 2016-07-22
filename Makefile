CHEFDK_VERSION=0.15.16-1
CHEFDK_DEB_NAME=chefdk_$(CHEFDK_VERSION)_amd64.deb
DOCKER_IMAGE_VERSION=$(CHEFDK_VERSION)-1
DOCKER_IMAGE_NAME=dappdeps/chefdk:$(DOCKER_IMAGE_VERSION)

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

all: $(HUB_IMAGE_FILE_PATH)

build/chefdk:
	@mkdir -p build
	wget https://packages.chef.io/stable/ubuntu/12.04/$(CHEFDK_DEB_NAME) -P build
	dpkg -x build/$(CHEFDK_DEB_NAME) build/chefdk

build/Dockerfile: build/chefdk
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	@echo "ADD chefdk /" >> build/Dockerfile

$(IMAGE_FILE_PATH): build/Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME) build
	@echo $(DOCKER_IMAGE_NAME) > $(IMAGE_FILE_PATH)

$(HUB_IMAGE_FILE_PATH): $(IMAGE_FILE_PATH)
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > $(HUB_IMAGE_FILE_PATH)

clean:
	@rm -rf build

.PHONY: all clean
