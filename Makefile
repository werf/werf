CHEFDK_VERSION=0.17.3

DOCKER_IMAGE_VERSION=$(CHEFDK_VERSION)-1
DOCKER_IMAGE_NAME=dappdeps/chefdk:$(DOCKER_IMAGE_VERSION)

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

all: $(HUB_IMAGE_FILE_PATH)

build/v$(CHEFDK_VERSION).tar.gz:
	@wget https://github.com/chef/chef-dk/archive/v$(CHEFDK_VERSION).tar.gz -P build

build/chefdk_$(CHEFDK_VERSION).deb: build/v$(CHEFDK_VERSION).tar.gz
	@tar xf build/v$(CHEFDK_VERSION).tar.gz -C build
	@echo 'install_dir "/.dapp/deps/chefdk"' >> build/chef-dk-$(CHEFDK_VERSION)/omnibus_overrides.rb
	@sed -i -e 's/install_dir: \/opt\/chefdk/install_dir: \/.dapp\/deps\/chefdk/g' build/chef-dk-$(CHEFDK_VERSION)/omnibus/.kitchen.yml
	@sed -i -e 's/INSTALLER_DIR=\/opt\/chefdk/INSTALLER_DIR=\/.dapp\/deps\/chefdk/g' build/chef-dk-$(CHEFDK_VERSION)/omnibus/package-scripts/chefdk/postinst
	@rm -f build/chef-dk-$(CHEFDK_VERSION)/omnibus/pkg/chefdk_$(CHEFDK_VERSION)*.deb
	@docker run -ti --rm --volume `pwd`:/app ubuntu:14.04 bash -ec '\
		apt-get update -qq && \
		apt-get install -y git autoconf build-essential devscripts curl && \
		git config --global user.email $(shell git config --global user.email) && \
		git config --global user.name $(shell git config --global user.name) && \
		gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3 && \
		curl -sSL https://get.rvm.io | bash && \
		source /etc/profile.d/rvm.sh && \
		rvm install 2.3.1 && \
		gem install bundler && \
		cd /app/build/chef-dk-$(CHEFDK_VERSION)/omnibus && \
		bundle install --without development && \
		bundle exec omnibus build -o install_dir:/.dapp/deps/chefdk -o append_timestamp:false chefdk'
	@cp build/chef-dk-$(CHEFDK_VERSION)/omnibus/pkg/chefdk_$(CHEFDK_VERSION)-1_amd64.deb \
      build/chefdk_$(CHEFDK_VERSION).deb

build/chefdk_$(CHEFDK_VERSION): build/chefdk_$(CHEFDK_VERSION).deb
	dpkg -x build/chefdk_$(CHEFDK_VERSION).deb build/chefdk_$(CHEFDK_VERSION)

build/Dockerfile_$(DOCKER_IMAGE_VERSION): build/chefdk_$(CHEFDK_VERSION)
	@echo "FROM scratch" > build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	@echo "CMD [\"no_such_command\"]" >> build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	@echo "ADD chefdk_$(CHEFDK_VERSION) /" >> build/Dockerfile_$(DOCKER_IMAGE_VERSION)

$(IMAGE_FILE_PATH): build/Dockerfile_$(DOCKER_IMAGE_VERSION)
	docker build -t $(DOCKER_IMAGE_NAME) -f build/Dockerfile_$(DOCKER_IMAGE_VERSION) build
	@echo $(DOCKER_IMAGE_NAME) > $(IMAGE_FILE_PATH)

$(HUB_IMAGE_FILE_PATH): $(IMAGE_FILE_PATH)
	docker push $(DOCKER_IMAGE_NAME)
	@echo $(DOCKER_IMAGE_NAME) > $(HUB_IMAGE_FILE_PATH)

clean:
	@rm -rf build

.PHONY: all clean
