setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "become_user module perform without errors (FIXME https://github.com/flant/werf/issues/1806)" {
	skip
	werf build -s :local
}
