CHEFDK_VERSION=0.15.16-1
CHEFDK_DEB=chefdk_$(CHEFDK_VERSION)_amd64.deb

all: image

build/$(CHEFDK_DEB):
	@mkdir -p build
	wget https://packages.chef.io/stable/ubuntu/12.04/$(CHEFDK_DEB) -P build
	dpkg -x build/$(CHEFDK_DEB) build/chefdk

build/Dockerfile: build/$(CHEFDK_DEB)
	@echo "FROM scratch" > build/Dockerfile
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile
	@echo "ADD chefdk /" >> build/Dockerfile

image: build/Dockerfile
	docker build -t dapp-chefdk build

clean:
	@rm -rf build

.PHONY: all clean image
