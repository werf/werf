#!/bin/bash

set -e

shopt -s nocasematch
# artifact  releases.json
# tag_name, html_url, body
PER_PAGE=100
PAGES=1
link_re='^.*<http.+[\&\?]page=([0-9]+)[^;]*; rel="last"'
if [[ "$(curl -Is  'https://api.github.com/repos/werf/werf/releases?per_page=${PER_PAGE}' | grep link | sed 's/^link: //')" =~ $link_re ]]; then
  PAGES=${BASH_REMATCH[1]}
fi

echo "There are ${PAGES} pages (${PER_PAGE} items per page) to get."

(for (( PAGE=1 ; $PAGE <= $PAGES; PAGE++ )); do
    curl -u werfio:${WERFIO_GITHUB_TOKEN} -s  "https://api.github.com/repos/werf/werf/releases?per_page=${PER_PAGE}&page=${PAGE}" | jq -cM '.|  map(select( (.tag_name|test("^v")) and (.tag_name | test("^v1.0") | not ) )) | map(.body = "{% raw %}\(.body | gsub("\n";"  \n") ){% endraw %}\n\n") | .[] | {tag_name: .tag_name, html_url: .html_url, body: .body} '
    echo "Got page ${PAGE}." >&2
done) | jq -sc ' . | {releases: .} ' > releases.json

echo "Finish getting pages."

RELEASES_COUNT=$(jq '.releases | length' releases.json)
if [ $(( $RELEASES_COUNT )) -lt 3 ]; then echo "Too low releases count - ${RELEASES_COUNT:-0} (seems like github rate-limit..)." ; exit 1; fi

HISTORY_JSON=${1:-git_history.json}

bash ./get_git_history.sh | jq -s '{"history":.}' > $HISTORY_JSON

./convert $HISTORY_JSON > releases_history.json

(for i in $(cat $HISTORY_JSON | jq '.history| unique_by(.group) | .[].group' | xargs); do
  echo "$(cat $HISTORY_JSON | jq --arg group $i '.history | map(select (.group == $group)) | .[0] as $item | $item.channels[] | select((.version | length) >0) | {"\(.name)-\($group)": .version}')"
done  ) | jq -s '{"versions": .|add } ' 1> channels_versions.json

# make archive with feeds
rm -rf feeds 2>/dev/null
mkdir -p feeds feeds/pages_en feeds/pages_ru
for group in $(cat $HISTORY_JSON | jq -r '.history| unique_by(.group) | .[].group' | xargs); do
  for channel in alpha beta ea stable rock-solid; do
    sed "s/##GROUP##/${group}/" feed-group-channel.xml-template | sed "s/##CHANNEL##/${channel}/" > feeds/pages_en/feed-${group}-${channel}.xml
    sed "s/##GROUP##/${group}/" feed-group-channel-ru.xml-template | sed "s/##CHANNEL##/${channel}/" > feeds/pages_ru/feed-${group}-${channel}.xml
  done
done

tar czf feeds.tgz feeds

base64 feeds.tgz > feeds.tgz.base64
rm -rf feeds feeds.tgz 2>/dev/null

rm -f $HISTORY_JSON
