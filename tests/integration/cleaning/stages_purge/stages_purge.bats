load ../../../helpers/common

setup() {
    werf_home_init
    test_dir_create
    test_dir_cd
}

teardown() {
    test_dir_werf_stages_purge
    test_dir_rm
    werf_home_deinit
}

stages_count() {
    docker images werf-stages-storage/werf-test-stages-purge -q | wc -l
}

@test "stages purge" {
    export WERF_STAGES_STORAGE=:local

    cp $BATS_TEST_DIRNAME/data/werf.yaml .

    werf build
    [ ! "$(stages_count)" -eq "0" ]

    werf stages purge
    [ "$(stages_count)" -eq "0" ]
}
