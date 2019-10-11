load ../../../../../helpers/common

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

files_checksum_command() {
    echo "find ${1:-.} -xtype f -not -path '**/.git' -not -path '**/.git/*' | xargs md5sum | awk '{ print \$1 }' | sort | md5sum | awk '{ print \$1 }'"
}

files_checksum() {
    eval "$(files_checksum_command ${1:-.})"
}

container_files_checksum() {
    image_name=$(werf run -s :local --dry-run | tail -n1 | cut -d' ' -f3)
    docker run --rm $image_name bash -ec "eval $(files_checksum_command ${1:-/app})"
}

@test "gitArchive, gitCache and gitLatestPatch stages" {
    git init

    # check: file werf.yaml is added on gitArchive stage

    cat << EOF > werf.yaml
project: werf-test-stapel-image-git-stages-base
configVersion: 1
---
image: ~
from: ubuntu
git:
- to: /app
EOF
    git add werf.yaml
    git commit -m "Initial commit"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "gitCache:               <empty>" ]]
    [[ "$output" =~ "gitLatestPatch:         <empty>" ]]
    [[ "$output" =~ "Git files will be actualized on stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitArchive" ]]

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "Building stage " ]]

    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app)" ]]

    # check: file test is added on gitLatestPatch stage

    date > test
    git add test
    git commit -m "Add file"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "gitCache:               <empty>" ]]
    [[ "$output" =~ "Git files will be actualized on stage gitLatestPatch" ]]
    [[ "$output" =~ "Use cache image for stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitLatestPatch" ]]

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "Building stage " ]]

    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app)" ]]

    # check: file large is added on gitCache stage

    openssl rand -base64 $((1024*1024)) > large
    git add large
    git commit -m "Add large file"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "gitLatestPatch:         <empty>" ]]
    [[ "$output" =~ "Git files will be actualized on stage gitCache" ]]
    [[ "$output" =~ "Use cache image for stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitCache" ]]

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "Building stage " ]]

    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app)" ]]

    # check: files are added on gitArchive stage (reset commit [werf reset]|[reset werf])

    git commit --allow-empty -m "[werf reset] Reset git archive"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "gitCache:               <empty>" ]]
    [[ "$output" =~ "gitLatestPatch:         <empty>" ]]
    [[ "$output" =~ "Git files will be actualized on stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitArchive" ]]

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "Building stage " ]]

    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app)" ]]
}
