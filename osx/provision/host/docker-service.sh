#!/bin/bash

set -e

sudo apt-get install -y docker-ce

sudo sed -i -e "s@^ExecStart=/usr/bin/dockerd -H fd://\$@ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2375@" /lib/systemd/system/docker.service

sudo systemctl daemon-reload
sudo systemctl restart docker
