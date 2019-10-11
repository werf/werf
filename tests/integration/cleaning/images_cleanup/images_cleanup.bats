load ../../../helpers/common

setup() {
    werf_home_init
    docker_registry_run
    test_dir_create
    test_dir_cd

    git init --bare remote.git
    git init
    git remote add origin remote.git

    export WERF_IMAGES_REPO=$WERF_TEST_DOCKER_REGISTRY/test
    export WERF_STAGES_STORAGE=:local
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

@test "images cleanup: git branch strategy" {
    cat << EOF > werf.yaml
project: werf-test-images-cleanup-branch
configVersion: 1
---
image: ~
from: alpine
shell:
  setup: date
EOF

    git add werf.yaml
    git commit -m "Initial commit"

    # check: command removes image that associated with local branch

    git checkout -b branchX
    werf build-and-publish --tag-git-branch branchX
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "branchX" ]]

    # check: command does not remove image that associated with remote branch

    git push --set-upstream origin branchX
    werf build-and-publish --tag-git-branch branchX
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ "$output" =~ "branchX" ]]

    # check: command removes image that associated with non-existing remote branch

    git checkout master
    git push origin --delete branchX
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "branchX" ]]
}

@test "images cleanup: git tag strategy" {
    cat << EOF > werf.yaml
project: werf-test-images-cleanup-tag
configVersion: 1
---
image: ~
from: alpine
shell:
  setup: date
EOF

    git add werf.yaml
    git commit -m "Initial commit"

    # check: command does not remove image that associated with local tag

    git tag tagX
    werf build-and-publish --tag-git-tag tagX
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ "$output" =~ "tagX" ]]

    # check: command removes image that associated with non-existing local tag

    git tag -d tagX
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "tagX" ]]

    # check: command removes image based on WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS

    git tag tagX
    werf build-and-publish --tag-git-tag tagX
    WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=0 werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "tagX" ]]

    # check: command removes images based on WERF_GIT_TAG_STRATEGY_LIMIT

    git tag tagA
    git tag tagB
    git tag tagC
    werf build-and-publish --tag-git-tag tagA --tag-git-tag tagB --tag-git-tag tagC
    WERF_GIT_TAG_STRATEGY_LIMIT=1 werf images cleanup --without-kube
    [[ "$(repo_images_count)" -eq "1" ]]

    # TODO: check: command does not remove image that associated with remote tag

    # git push --tags
    # werf build-and-publish --tag-git-tag tagX
    # werf images cleanup --without-kube
    # run crane ls $WERF_IMAGES_REPO
    # [ "$status" -eq 0 ]
    # [[ "$output" =~ "tagX" ]]

    # TODO: check: command removes image that associated with non-existing remote branch

    # git push origin --delete branchX
    # werf images cleanup --without-kube
    # run crane ls $WERF_IMAGES_REPO
    # [ "$status" -eq 0 ]
    # [[ ! "$output" =~ "branchX" ]]
}

@test "images cleanup: git commit strategy" {
    cat << EOF > werf.yaml
project: werf-test-images-cleanup-commit
configVersion: 1
---
image: ~
from: alpine
shell:
  setup: date
EOF

    git add werf.yaml
    git commit -m "Initial commit"

    # check: command removes image that associated with non-existing commit

    werf build-and-publish --tag-git-commit 8a99331ce0f918b411423223f4060e9688e03f6a
    werf images cleanup --without-kube
    [[ "$(repo_images_count)" -eq "0" ]]

    # check: command does not remove image that associated with local commit

    commit=$(git rev-parse HEAD)
    werf build-and-publish --tag-git-commit $commit
    werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ "$output" =~ "$commit" ]]

    # check: command removes image based on WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS

    WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS=0 werf images cleanup --without-kube
    run crane ls $WERF_IMAGES_REPO
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "$commit" ]]

    # check: command removes image based on WERF_GIT_COMMIT_STRATEGY_LIMIT

    git commit --allow-empty --allow-empty-message -m ""
    werf build-and-publish --tag-git-commit $(git rev-parse HEAD)
    git commit --allow-empty --allow-empty-message -m ""
    werf build-and-publish --tag-git-commit $(git rev-parse HEAD)
    git commit --allow-empty --allow-empty-message -m ""
    werf build-and-publish --tag-git-commit $(git rev-parse HEAD)
    [[ "$(repo_images_count)" -eq "3" ]]

    WERF_GIT_COMMIT_STRATEGY_LIMIT=1 werf images cleanup --without-kube
    [[ "$(repo_images_count)" -eq "1" ]]
}
