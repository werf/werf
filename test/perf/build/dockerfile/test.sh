#!/usr/bin/env bash
set -xeuo pipefail

vanilla_build_enabled=0
native_build_enabled=0
werf_path="$HOME/go/bin/werf"
usage="Usage: $(basename "$0") [options] docker_repository

Options:
  -e  path to werf executable (default: $werf_path)
  -v  do vanilla build (default: $vanilla_build_enabled)
  -n  do native-rootless build (default: $native_build_enabled)

Example:
  $(basename "$0") -vfn docker.io/user/repo:20.04"

temp_dir="/tmp/werf-perf-test"
script_dir="$(
  cd "$(dirname "$0")"
  pwd
)"

function get_abs_path {
  echo "$(
    cd "$(dirname "$1")"
    pwd
  )/$(basename "$1")"
}

while getopts ":vfne:" opt; do
  case ${opt} in
    v)
      vanilla_build_enabled=1
      ;;
    n)
      native_build_enabled=1
      ;;
    e)
      werf_path="$(get_abs_path "$OPTARG")"
      ;;
    *)
      echo "$usage"
      exit 1
      ;;
  esac
done
shift $((OPTIND - 1))
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

docker pull ubuntu:20.04 || true
podman pull ubuntu:20.04 || true
buildah pull ubuntu:20.04 || true

export WERF_REPO="$repo"
export WERF_LOG_DEBUG=1
export WERF_DISABLE_AUTO_HOST_CLEANUP=1
export WERF_PERF_TEST_CONTAINER_RUNTIME=1

if [[ $vanilla_build_enabled == 1 ]]; then
  echo "Running vanilla werf build"
  "$werf_path" build | tee ../vanilla-build.log
  echo "Finished vanilla werf build"
fi

if [[ $native_build_enabled == 1 ]]; then
  echo "Running native rootless werf build"
  WERF_BUILDAH_MODE="native-rootless" "$werf_path" build | tee ../native-rootless-build.log
  echo "Finished native rootless werf build"
fi

grep --color=never -HE ' seconds' ../*.log
echo "Logs available at $temp_dir"
