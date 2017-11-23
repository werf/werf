name 'dappdeps-base'
maintainer 'Timofey Kirillov'
homepage 'https://github.com/flant/dappdeps-base'

license 'MIT'
license_file 'LICENSE.txt'

DOCKER_IMAGE_VERSION = "0.2.1"

install_dir "/.dapp/deps/base/#{DOCKER_IMAGE_VERSION}"

build_version DOCKER_IMAGE_VERSION
build_iteration 1

dependency 'dappdeps-base'
