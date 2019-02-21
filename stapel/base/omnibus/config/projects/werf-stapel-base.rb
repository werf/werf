name "werf-stapel-base"
maintainer "Flant"
homepage "https://github.com/flant/werf"

license "Apache 2.0"
license_file "LICENSE"

DOCKER_IMAGE_VERSION = "0.3.0"

install_dir "/.werf/stapel/base/#{DOCKER_IMAGE_VERSION}"

build_version DOCKER_IMAGE_VERSION
build_iteration 1

dependency "werf-stapel-base"
