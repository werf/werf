load ../../../../helpers/common
load helpers

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

@test "Git patch apply should remove empty git directories (base)" {
    git init

    cat <<EOF > werf.yaml
project: werf-test-stapel-image-git-remove-empty-dirs-b
configVersion: 1
---
image: ~
from: ubuntu
git:
- to: /app
EOF

    git add werf.yaml
    git commit -m "Initial commit"

    dirname=$(quote_shell_arg "dir/sub dir/sub dir with special ch@ra(c)ters? ()")
    filename=$dirname/file
    eval $(echo mkdir -p $dirname)
    eval $(echo echo "test > $filename")
    eval $(echo git add $filename)
    git commit -m "+"

    werf build --stages-storage :local
    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app)" ]]

    eval $(echo git rm $filename)
    git commit -m "-"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Use cache image for stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitLatestPatch" ]]

    image_name=$(werf run -s :local --dry-run | tail -n1 | cut -d' ' -f3)
    run docker run --rm $image_name test -d /app/dir
    [ "$status" -eq 1 ]
}

@test "Git patch apply should remove empty git directories (except user directory)" {
    git init

    cat <<EOF > werf.yaml
project: werf-test-stapel-image-git-remove-empty-dirs-eud
configVersion: 1
---
image: ~
from: ubuntu
git:
- to: /app
shell:
  setup: mkdir '/app/dir/user dir'
EOF

    git add werf.yaml
    git commit -m "Initial commit"

    dirname=$(quote_shell_arg "dir/sub dir/sub dir with special ch@ra(c)ters? ()")
    filename=$dirname/file
    eval $(echo mkdir -p $dirname)
    eval $(echo echo "test > $filename")
    eval $(echo git add $filename)
    git commit -m "+"

    werf build --stages-storage :local
    [[ "$(files_checksum $WERF_TEST_DIR)" == "$(container_files_checksum /app "-not -path '/app/dir/user'")" ]]

    eval $(echo git rm $filename)
    git commit -m "-"

    run werf build --stages-storage :local
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Use cache image for stage gitArchive" ]]
    [[ "$output" =~ "Building stage gitLatestPatch" ]]

    image_name=$(werf run -s :local --dry-run | tail -n1 | cut -d' ' -f3)

    run docker run --rm $image_name bash -c "test -d $(quote_shell_arg '/app/dir/user dir')"
    [ "$status" -eq 0 ]

    run docker run --rm $image_name bash -c "test -d $(quote_shell_arg '/app/dir/sub dir')"
    [ "$status" -eq 1 ]
}
