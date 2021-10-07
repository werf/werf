#!/usr/bin/env bash
set -euo pipefail

usage="Usage: $(basename "$0") [options] docker_repository

Options:
  -v  do vanilla build
  -f  do docker-with-fuse build
  -n  do native-rootless-build

Example:
  $(basename "$0") -vfn docker.io/user/repo:20.04"

temp_dir="/tmp/werf-perf-test"
script_dir="$(cd "$(dirname "$0")"; pwd)"

vanilla_build_enabled=0
docker_with_fuse_build_enabled=0
native_rootless_build_enabled=0
while getopts ":vfn" opt; do
  case ${opt} in
    v )
      vanilla_build_enabled=1
      ;;
    f )
      docker_with_fuse_build_enabled=1
      ;;
    n )
      native_rootless_build_enabled=1
      ;;
    * )
      echo "$usage"
      exit 1
      ;;
  esac
done
shift $((OPTIND -1))
repo="$1"

rm -rf "$temp_dir"
mkdir -p "$temp_dir"
cd "$temp_dir"
git clone https://github.com/neovim/neovim neovim

cp -rf neovim neovim-v0.1.0
cp -rf neovim neovim-v0.5.0

find "$temp_dir/neovim" -mindepth 1 -maxdepth 1 -not -name .git -print0 | xargs -0 rm -rf

git -C neovim-v0.1.0 checkout v0.1.0
git -C neovim-v0.5.0 checkout v0.5.0

rm -rf neovim-v0.1.0/.git
rm -rf neovim-v0.5.0/.git

mv neovim-v0.1.0 neovim/
mv neovim-v0.5.0 neovim/

find "$script_dir" -mindepth 1 -maxdepth 1 -print0 | xargs -0 -I{} cp -rf '{}' neovim/

cd neovim
git init
git add .
git commit -m init

"${script_dir}/../go-build.sh"

docker pull ubuntu:20.04
podman pull ubuntu:20.04 || true
buildah pull ubuntu:20.04 || true

export WERF_REPO="$repo"
export WERF_LOG_DEBUG=1
export WERF_DISABLE_AUTO_HOST_CLEANUP=1

if [[ $vanilla_build_enabled == 1 ]]; then
  echo "Running vanilla werf build"
  ~/go/bin/werf build | tee ../vanilla-build.log
  echo "Finished vanilla werf build"
fi

if [[ $docker_with_fuse_build_enabled == 1 ]]; then
  echo "Running docker-with-fuse werf build"
  WERF_CONTAINER_RUNTIME_BUILDAH="docker-with-fuse" ~/go/bin/werf build | tee ../docker-with-fuse-build.log
  echo "Finished docker-with-fuse werf build"
fi

if [[ $native_rootless_build_enabled == 1 ]]; then
  echo "Running native-rootless werf build"
  WERF_CONTAINER_RUNTIME_BUILDAH="native-rootless" ~/go/bin/werf build | tee ../native-rootless-build.log
  echo "Finished native-rootless werf build"
fi

grep --color=never -HE ' seconds' ../*.log
echo "Logs available at $temp_dir"
