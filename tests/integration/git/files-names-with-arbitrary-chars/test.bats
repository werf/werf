setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf --dir app2 stages purge -s :local

	rm -rf repo
	rm app1/.git
	rm app2/.git
}

@test "Arbitrary chars in git files names should be allowed (FIXME https://github.com/flant/werf/issues/1711)" {
	skip

	git init app1 --separate-git-dir=repo
	git -C app1 add .
	git -C app1 commit -m app1
	werf --dir app1 build -s :local

	git init app2 --separate-git-dir=repo
	git -C app2 add .
	git -C app2 commit -m app2
	werf --dir app2 build -s :local
}
