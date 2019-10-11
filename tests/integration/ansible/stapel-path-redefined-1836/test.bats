setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "Non standard PATH should not be redefined in stapel build container (https://github.com/flant/werf/issues/1836)" {
	werf build -s :local
}

