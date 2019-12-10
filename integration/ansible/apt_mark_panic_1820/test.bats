setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "apt-mark from apt ansible module should not panic in all supported ubuntu versions (https://github.com/flant/werf/issues/1820)" {
	werf build -s :local
}

