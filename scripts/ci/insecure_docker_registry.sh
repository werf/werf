#!/bin/bash -e

sudo bash -c "cat <<EOF > /etc/docker/daemon.json
{ \"insecure-registries\":[\"$WERF_TEST_K8S_INSECURE_DOCKER_REGISTRY\"] }
EOF"

sudo systemctl daemon-reload
sudo service docker restart
