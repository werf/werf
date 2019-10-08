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

stages_count() {
    docker images werf-stages-storage/werf-test-stages-cleanup -q | wc -l
}

@test "stages cleanup" {
    export WERF_IMAGES_REPO=$WERF_TEST_DOCKER_REGISTRY/test
    export WERF_STAGES_STORAGE=:local

    cp $BATS_TEST_DIRNAME/data/werf.yaml .

    # check: cleanup stages storage based on empty images repository
    werf stages cleanup

    # build and publish application first time

    export FROM_CACHE_VERSION=1
    werf build-and-publish --tag-custom test
    stages_amount_1=$(stages_count)

    # check: cleanup does not delete any stages

    werf stages cleanup
    [ "$stages_amount_1" -eq "$(stages_count)" ]

    # fully rebuild and publish application

    export FROM_CACHE_VERSION=2
    werf build-and-publish --tag-custom test
    stages_amount_2=$(stages_count)

    # check: cleanup does not delete any stages (stages cannot be removed during 2 hours by specific policy)

    werf stages cleanup
    [ "$stages_amount_2" -eq "$(stages_count)" ]

    # turn off policy
    export WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY=true

    # check: cleanup deletes stages that do not have related images in images repository

    werf stages cleanup
    [ "$(($stages_amount_2-$stages_amount_1))" -eq "$(stages_count)" ]
}
