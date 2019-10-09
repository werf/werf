load ../../../../helpers/common

setup() {
    werf_home_init
    test_dir_create
    test_dir_cd
}

teardown() {
    docker rm --force $CONTAINER_NAME
    test_dir_werf_stages_purge
    test_dir_rm
    werf_home_deinit
}

@test "artifacts (FIXME https://github.com/flant/werf/issues/1820)" {
    skip
    cp -r $BATS_TEST_DIRNAME/data/werf.yaml .

    werf build --stages-storage :local

    CONTAINER_NAME=go-booking-artifacts-$(generate_random_string)
    container_host_port=$(get_unused_port)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d -p $container_host_port:9000 --name $CONTAINER_NAME" go-booking -- /app/run.sh

    wait_till_host_ready_to_respond localhost:$container_host_port
    run curl localhost:$container_host_port
    [ "$status" -eq 0 ]
    [[ "$output" =~ "revel framework booking demo" ]]

    docker stop $CONTAINER_NAME
}
