setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "Ansible successfully install packages into centos image using yum module (FIXME https://github.com/flant/werf/issues/1798)" {
	skip
	werf build -s :local
}
