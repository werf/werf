CHEFDK_VERSION=0.15.16-1
CHEFDK_DEB=chefdk_$(CHEFDK_VERSION)_amd64.deb
CHEFDK_LOCAL_IMAGE=dapp-chefdk:$(CHEFDK_VERSION)
CHEFDK_IMAGE=dapp/chefdk:$(CHEFDK_VERSION)

all: build/hub_image

build/chefdk:
	@mkdir -p build
	wget https://packages.chef.io/stable/ubuntu/12.04/$(CHEFDK_DEB) -P build
	dpkg -x build/$(CHEFDK_DEB) build/chefdk

build/Dockerfile: build/chefdk
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	@echo "ADD chefdk /" >> build/Dockerfile

build/image: build/Dockerfile
	docker build -t $(CHEFDK_LOCAL_IMAGE) build
	@echo $(CHEFDK_LOCAL_IMAGE) > build/image

build/hub_image: build/image
	@echo push to dapp-chefdk
	@echo $(CHEFDK_IMAGE) > build/hub_image

clean:
	@rm -rf build

.PHONY: all clean image
