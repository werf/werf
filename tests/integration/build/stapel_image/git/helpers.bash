#!/bin/bash

files_checksum_command() {
    echo -n "find ${1:-.} -xtype f -not -path '**/.git' -not -path '**/.git/*' ${@:2} -exec bash -c 'printf \"%s\n\" \"\${@@Q}\"' sh {} + | xargs md5sum | awk '{ print \$1 }' | sort | md5sum | awk '{ print \$1 }'" | base64 -w 0
}

files_checksum() {
    eval "$(files_checksum_command ${1:-.} ${@:2} | base64 -d -w 0)"
}

container_files_checksum() {
    image_name=$(werf run -s :local --dry-run | tail -n1 | cut -d' ' -f3)
    cmd=$(printf "docker run --rm %s bash -ec 'eval \$(echo -n %s | base64 -d -w 0)'" $image_name $(files_checksum_command ${1:-/app} ${@:2}))
    eval "$cmd"
}

quote_shell_arg() {
  printf "%s\n" "${@@Q}"
}
