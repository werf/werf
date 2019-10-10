load ../../../helpers/common

setup() {
    werf_home_init
    docker_registry_run
    test_dir_create
    test_dir_cd
}

teardown() {
    test_dir_werf_stages_purge
    test_dir_rm
    docker_registry_rm
    werf_home_deinit
}

repo_images_count() {
    crane ls $WERF_IMAGES_REPO | wc -l
}

@test "images purge" {
    export WERF_STAGES_STORAGE=:local
    export WERF_IMAGES_REPO=$WERF_TEST_DOCKER_REGISTRY/test

    cp $BATS_TEST_DIRNAME/data/werf.yaml .

    # check: purge empty images repository without error
    werf images purge

    # check: command ignores images which are not produced with werf

    docker tag alpine $WERF_IMAGES_REPO:test
    docker push $WERF_IMAGES_REPO:test
    werf images purge
    [ "$(repo_images_count)" -eq "1" ]

    # check: command deletes all werf images

    werf build-and-publish --tag-custom a --tag-custom b --tag-custom c
    [ "$(repo_images_count)" -eq "4" ]

    werf images purge
    [ "$(repo_images_count)" -eq "1" ]
}
