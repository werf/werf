load ../../../../helpers/common
load first_application

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
