load ../../helpers/common

setup() {
    werf_home_init
    test_dir_create
    test_dir_cd
}

teardown() {
    test_dir_rm
    werf_home_deinit
}

@test "secret key" {
    test_data_secret_path=$BATS_TEST_DIRNAME/data/secret

    run bash -c "$test_data_secret_path | werf helm secret encrypt"
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Error: encryption key not found in" ]]

    export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

    cat $test_data_secret_path | werf helm secret encrypt
}

@test "secret file" {
    test_data_secret_path=$BATS_TEST_DIRNAME/data/secret
    encrypted_secret_path=.helm/secret/encrypted_secret

    export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

    # check: secret file encryption

    werf helm secret file encrypt $test_data_secret_path -o $encrypted_secret_path
    run diff --brief <(sort $encrypted_secret_path) <(sort $test_data_secret_path) >/dev/null
    [ "$status" -eq 1 ]

    # check: secret file decryption

    werf helm secret file decrypt $encrypted_secret_path -o secret
    run diff --brief <(sort secret) <(sort $test_data_secret_path) >/dev/null
    [ "$status" -eq 0 ]

    # check: secret file rotation

    export WERF_OLD_SECRET_KEY=$WERF_SECRET_KEY
    export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

    werf helm secret rotate-secret-key
    werf helm secret file decrypt $encrypted_secret_path
}

@test "secret values" {
    test_data_secret_yaml_path=$BATS_TEST_DIRNAME/data/secret.yaml

    export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

    # check: secret values encryption

    werf helm secret values encrypt $test_data_secret_yaml_path -o encrypted_secret.yaml
    run diff --brief <(sort encrypted_secret.yaml) <(sort $test_data_secret_yaml_path) >/dev/null
    [ "$status" -eq 1 ]

    # check: secret values decryption

    werf helm secret values decrypt encrypted_secret.yaml -o secret.yaml
    run diff --brief <(sort secret.yaml) <(sort $test_data_secret_yaml_path) >/dev/null
    [ "$status" -eq 0 ]

    # check: secret values rotation

    export WERF_OLD_SECRET_KEY=$WERF_SECRET_KEY
    export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

    werf helm secret rotate-secret-key encrypted_secret.yaml
    werf helm secret values decrypt encrypted_secret.yaml
}
