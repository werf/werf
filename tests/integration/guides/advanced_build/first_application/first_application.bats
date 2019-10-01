load ../../../../helpers/common

setup() {
    werf_home_init
    docker_registry_run
    test_dir_create
    test_dir_cd
}

teardown() {
    docker rm --force $CONTAINER_NAME
    docker_registry_rm
    test_dir_werf_stages_purge
    test_dir_rm
    werf_home_deinit
}

test_case_run() {
  werf build --stages-storage :local

  container_name=$1
  container_host_port=$(get_unused_port)
  werf run \
    --stages-storage :local \
    --docker-options="--rm -d -p $container_host_port:8000 --name $container_name" -- /app/start.sh

  wait_till_host_ready_to_respond localhost:$container_host_port
  run curl localhost:$container_host_port
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Symfony Demo application" ]]

  registry_repository_name=$container_name
  werf publish \
    --stages-storage :local \
    --images-repo $WERF_TEST_DOCKER_REGISTRY/$registry_repository_name \
    --tag-custom v0.1.0

  docker stop $container_name
}

@test "first application with ansible" {
    git clone https://github.com/symfony/symfony-demo.git .
    cp -r $BATS_TEST_DIRNAME/data/ansible/* .

    CONTAINER_NAME=symfony-demo-ansible-$(generate_random_string)
    test_case_run $CONTAINER_NAME
}

@test "first application with shell" {
    git clone https://github.com/symfony/symfony-demo.git .
    cp -r $BATS_TEST_DIRNAME/data/shell/* .

    CONTAINER_NAME=symfony-demo-shell-$(generate_random_string)
    test_case_run $CONTAINER_NAME
}
