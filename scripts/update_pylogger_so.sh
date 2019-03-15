#!/bin/bash

go build -o pkg/build/builder/ansible/pylogger.so -buildmode=c-shared github.com/flant/werf/pkg/pylogger
