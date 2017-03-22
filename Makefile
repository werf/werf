DOCKER_IMAGE_VERSION = $(shell cat config/projects/dappdeps-base.rb | \
                               grep 'DOCKER_IMAGE_VERSION =' | \
                               grep -oP '[0-9.]+')
DOCKER_IMAGE_NAME=dappdeps/base:$(DOCKER_IMAGE_VERSION)

IMAGE_FILE_PATH=build/image_$(DOCKER_IMAGE_VERSION)
HUB_IMAGE_FILE_PATH=build/hub_image_$(DOCKER_IMAGE_VERSION)

BUILDENV_DOCKER_IMAGE=centos:5

all: $(HUB_IMAGE_FILE_PATH)

build/dappdeps-base_$(DOCKER_IMAGE_VERSION).rpm:
	@rm -f pkg/dappdeps-base-$(DOCKER_IMAGE_VERSION)*.rpm
	@docker run --rm --volume `pwd`:/app $(BUILDENV_DOCKER_IMAGE) bash -ec '\
		yum install -y epel-release.noarch && \
		yum install -y make gpg git curl which file gettext-devel libattr-devel sudo man unzip gcc-c++ screen rpm-build libtermcap && \
		mkdir -p /usr/src/redhat/SOURCES && \
		mkdir -p /usr/src/redhat/SPECS && \
		cp /app/centos-extras/tar.spec /usr/src/redhat/SPECS/tar.spec && \
		curl -sR -o /usr/src/redhat/SOURCES/tar-1.26.tar.bz2 ftp://ftp.gnu.org/gnu/tar/tar-1.26.tar.bz2 && \
		rpmbuild -ba /usr/src/redhat/SPECS/tar.spec && \
		rpm -ivh --replacefiles --replacepkgs --oldpackage /usr/src/redhat/RPMS/x86_64/tar-1.26-1.x86_64.rpm && \
		git config --global user.email $(shell git config --global user.email) && \
		git config --global user.name $(shell git config --global user.name) && \
		gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3 && \
		curl -sSL https://get.rvm.io | bash && \
		source /etc/profile.d/rvm.sh && \
		rvm install 2.3.1 && \
		gem install bundler && \
		cd /app && \
		bundle install --without development && \
		bundle exec omnibus build -o append_timestamp:false dappdeps-base'
	@cp pkg/dappdeps-base-$(DOCKER_IMAGE_VERSION)-1.el5.x86_64.rpm \
      build/dappdeps-base_$(DOCKER_IMAGE_VERSION).rpm

build/dappdeps-base_$(DOCKER_IMAGE_VERSION): build/dappdeps-base_$(DOCKER_IMAGE_VERSION).rpm
	mkdir build/dappdeps-base_$(DOCKER_IMAGE_VERSION)
	cd build/dappdeps-base_$(DOCKER_IMAGE_VERSION) && rpm2cpio ../dappdeps-base_$(DOCKER_IMAGE_VERSION).rpm | cpio -idmv

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
