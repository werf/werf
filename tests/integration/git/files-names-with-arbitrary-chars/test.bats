setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge --dir app2 -s :local
	rm -rf repo app1 app2
}

@test "Arbitrary chars in git files names should be allowed (FIXME https://github.com/flant/werf/issues/1711)" {
	skip

	tar xf app.tar.gz

	git init app1 --separate-git-dir=repo
	git -C app1 add .
	git -C app1 commit -m app1
	werf build --dir app1 -s :local

	git init app2 --separate-git-dir=repo
	git -C app2 add .
	git -C app2 commit -m app2
	werf build --dir app2 -s :local
}
