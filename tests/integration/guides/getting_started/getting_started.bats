load ../../../helpers/common

setup() {
    werf_home_init
    docker_registry_run
    test_dir_create
    test_dir_cd
}

teardown() {
    docker rm --force $CONTAINER_NAME
    test_dir_werf_stages_purge
    test_dir_rm
    docker_registry_rm
    werf_home_deinit
}

@test "getting started" {
    git clone https://github.com/dockersamples/linux_tweet_app.git .
    cp -r $BATS_TEST_DIRNAME/data/werf.yaml .

    werf build --stages-storage :local

    CONTAINER_NAME=getting_started-$(generate_random_string)
    container_host_port=$(get_unused_port)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d -p $container_host_port:80 --name $CONTAINER_NAME"

    run curl localhost:$container_host_port
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Linux Tweet App!" ]]

    docker stop $CONTAINER_NAME

    registry_repository_name=$CONTAINER_NAME
    werf publish \
        --stages-storage :local \
        --images-repo $WERF_TEST_DOCKER_REGISTRY/$registry_repository_name \
        --tag-custom v0.1.0
}
