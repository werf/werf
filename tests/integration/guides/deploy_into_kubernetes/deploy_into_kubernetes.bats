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

@test "deploy into kubernetes" {
    skip "This command will return zero soon, but not now"

    cp -r $BATS_TEST_DIRNAME/data/* .

    werf build-and-publish \
        --stages-storage :local \
        --images-repo :minikube \
        --tag-custom myapp

    werf deploy \
        --stages-storage :local \
        --images-repo :minikube \
        --tag-custom myapp \
        --env dev

    werf dismiss \
        --env dev \
        --with-namespace
}
