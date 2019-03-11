name "werf-stapel-ansible"
maintainer "Flant"
homepage "https://github.com/flant/werf"

license "Apache 2.0"
license_file "LICENSE"

DOCKER_IMAGE_VERSION = "2.4.4.0-11"

install_dir "/.werf/stapel/ansible/#{DOCKER_IMAGE_VERSION}"

build_version DOCKER_IMAGE_VERSION
build_iteration 1

dependency "werf-stapel-ansible"
