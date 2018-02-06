name "dappdeps-ansible"
maintainer "Timofey Kirillov"
homepage "https://github.com/flant/dappdeps-ansible"

license "MIT"
license_file "LICENSE.txt"

DOCKER_IMAGE_VERSION = "2.4.1.0-1"

install_dir "/.dapp/deps/ansible/#{DOCKER_IMAGE_VERSION}"

build_version DOCKER_IMAGE_VERSION
build_iteration 1

dependency "dappdeps-ansible"
