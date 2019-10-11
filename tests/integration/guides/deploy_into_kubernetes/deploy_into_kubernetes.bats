load ../../../helpers/common
load ../../../helpers/k8s

setup() {
    werf_home_init
    test_dir_create
    test_dir_cd
}

teardown() {
    werf dismiss \
      --env dev \
      --with-namespace

    test_dir_werf_stages_purge
    test_dir_rm
    werf_home_deinit
}

@test "[k8s] deploy into kubernetes$(test_k8s_version)" {
    test_skip_if_k8s_disabled
    test_requires_k8s_docker_registry

    registry_repository_name=deploy-into-kubernetes-$(generate_random_string)
    cp -a $BATS_TEST_DIRNAME/data/. .

    werf build-and-publish \
        --stages-storage :local \
        --images-repo $WERF_TEST_K8S_DOCKER_REGISTRY/$registry_repository_name \
        --tag-custom myapp

    werf deploy \
        --stages-storage :local \
        --images-repo $WERF_TEST_K8S_DOCKER_REGISTRY/$registry_repository_name \
        --tag-custom myapp \
        --env dev
}
