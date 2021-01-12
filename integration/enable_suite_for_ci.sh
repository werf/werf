#!/bin/bash

set -e

export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export project_dir=$script_dir/..

declare -a SUPPORTED_ci_layout_argS=($(ls ${project_dir}/integration/ci_suites))

supported_ci_layout_arg_usage_str="${SUPPORTED_ci_layout_argS[0]}"
for i in "${!SUPPORTED_ci_layout_argS[@]}" ; do
    if (( ${i} > 0 )) ; then
        supported_ci_layout_arg_usage_str="${supported_ci_layout_arg_usage_str}|${SUPPORTED_ci_layout_argS[$i]}"
    fi
done
USAGE="Usage:\n\t$0 SUITE_DIR {$supported_ci_layout_arg_usage_str}"

if [ "$1" == "" ] ; then
    echo -e "$USAGE" >&1
    echo >&1
    echo "Error: provide path to suite dir!" >&1
    exit 1    
fi
export path_arg="${1}"

if [[ ! " ${SUPPORTED_ci_layout_argS[@]} " =~ " ${2} " ]]; then
    echo -e "$USAGE" >&1
    echo >&1
    echo "Error: provide one of the supported ci layout! Use \"default\" if not sure." >&1
    exit 1
fi
export ci_layout_arg="${2}"

export from="$(realpath ${path_arg} --relative-to ${project_dir}/integration/ci_suites/${ci_layout_arg})"
export to="${project_dir}/integration/ci_suites/${ci_layout_arg}"

echo "Creating symlink $(realpath -s ${to}/$(basename ${path_arg}) --relative-to $(pwd)) to ${path_arg}"

ln -fs "${from}" "${to}"
