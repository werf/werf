#!/bin/bash

werf_home_init() {
# TODO: all tests have to use common werf locks that are stored in the service directory, $WERF_HOME/service
#  export WERF_HOME=$BATS_TMPDIR/werf-test-home-$(generate_random_string)
#  mkdir $WERF_HOME
  /bin/true
}

werf_home_deinit() {
#  docker run \
#    --rm \
#    --volume $WERF_HOME:$WERF_HOME \
#    alpine \
#    rm -rf $WERF_HOME/*
#
#  rmdir $WERF_HOME
  /bin/true
}

test_dir_create() {
  WERF_TEST_DIR=$BATS_TMPDIR/werf_test-$(generate_random_string)
  mkdir $WERF_TEST_DIR
}

test_dir_cd() {
  cd $WERF_TEST_DIR
}

test_dir_rm() {
  rm -rf $WERF_TEST_DIR
}

test_dir_werf_stages_purge() {
  werf stages purge \
    --stages-storage :local \
    --force \
    --dir $WERF_TEST_DIR
}

docker_registry_run() {
  WERF_TEST_DOCKER_REGISTRY_CONTAINER_NAME=werf_test_docker_registry-$(generate_random_string)
  container_host_port=$(get_unused_port)
  docker run \
    -d -p $container_host_port:5000 \
    -e REGISTRY_STORAGE_DELETE_ENABLED=true \
    --name $WERF_TEST_DOCKER_REGISTRY_CONTAINER_NAME registry:2

  WERF_TEST_DOCKER_REGISTRY=localhost:$container_host_port
  wait_till_host_ready_to_respond $WERF_TEST_DOCKER_REGISTRY 30
}

docker_registry_rm() {
  docker rm --force $WERF_TEST_DOCKER_REGISTRY_CONTAINER_NAME
}

generate_random_string() {
  date +%s.%N | sha256sum | cut -c 1-10
}

get_unused_port() {
  comm -23 <(seq 49152 65535) <(ss -tan | awk '{print $4}' | cut -d':' -f2 | grep "[0-9]\{1,5\}" | sort | uniq) | shuf | head -n 1
}

wait_till_host_ready_to_respond() {
  attempt_counter=0
  url=$1
  max_attempts=${2:-10}

  until $(curl --output /dev/null --silent --head --fail $url); do
      if [ ${attempt_counter} -eq ${max_attempts} ]
      then
        echo "Max attempts reached" >&3
        exit 1
      fi

      attempt_counter=$(($attempt_counter+1))
      sleep 1
  done
}
