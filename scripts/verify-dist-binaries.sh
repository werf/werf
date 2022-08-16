#!/bin/bash
set -euo pipefail

script_dir="$(cd "$( dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
project_dir="$script_dir/.."

version="${1:?Version should be set}"

declare -A regexps
regexps["$project_dir/dist/$version/linux-amd64/bin/werf"]="x86-64.*statically linked.*Linux"
regexps["$project_dir/dist/$version/linux-arm64/bin/werf"]="ARM aarch64.*statically linked.*Linux"
regexps["$project_dir/dist/$version/darwin-amd64/bin/werf"]="Mach-O.*x86_64"
regexps["$project_dir/dist/$version/darwin-arm64/bin/werf"]="Mach-O.*arm64"
regexps["$project_dir/dist/$version/windows-amd64/bin/werf.exe"]="x86-64.*Windows"

for filename in "${!regexps[@]}"; do
  if ! [[ -f "$filename" ]]; then
    echo Binary at "$filename" does not exist.
    exit 1
  fi

  file "$filename" | awk -v regexp="${regexps[$filename]}" '{print $0; if ($0 ~ regexp) { exit } else { print "Unexpected binary info ^^"; exit 1 }}'
done
