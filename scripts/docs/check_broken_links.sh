#!/bin/bash

set -e

arg_site_lang="${1:?ERROR: Site language \'main\' or \'ru\' should be specified as the first argument.}"
arg_werf_bin_path="${2:?ERROR: werf bin path should be specified as the second argument.}"

source_path="$(realpath "${BASH_SOURCE[0]}")"
project_dir="$(dirname $source_path)/../.."
docs_dir="$project_dir/docs"

script=$(cat <<EOF
cd /app/docs && \
  bundle exec htmlproofer \
    --allow-hash-href \
    --empty-alt-ignore \
    --check_html \
    --url_ignore '/localhost/,/example.com/,/atseashop.com/,/https\:\/\/t.me/,/.slack.com/,/docs.github.com/,/help.github.com/,/habr.com/,/cncf.io/,/\/guides/,/\/how_it_works\.html/,/\/installation\.html/,/werf_yaml.html#configuring-cleanup-policies/,/css\/configuration-table.css/,/kubernetes.io\/cluster-service=true/' \
    --url_swap 'documentation/v[0-9]+[^/]+/:documentation/' \
    --http-status-ignore '0,429' \
    /app/_site/$arg_site_lang
EOF
)

$arg_werf_bin_path run --dir=$docs_dir assets --dev --docker-options="--entrypoint=bash" -- -c "$script"
