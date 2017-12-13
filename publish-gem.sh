#!/bin/bash

set -e

rm -f dapp-*.gem
gem build ./dapp.gemspec
gem push dapp-*.gem
