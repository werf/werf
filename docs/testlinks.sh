#!/bin/bash

bundle exec htmlproofer \
--allow-hash-href --empty-alt-ignore  --check_html \
--url_ignore "/localhost/,/example.com/,/atseashop.com/,/https\:\/\/t.me/,/.slack.com/,/cncf.io/,/\/guides\.html/,/\/introduction\.html/,/\/installation\.html/" _site/