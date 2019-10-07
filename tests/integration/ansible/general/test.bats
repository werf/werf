setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "Ansible modules should perform without errors" {
	werf build -s :local
}
