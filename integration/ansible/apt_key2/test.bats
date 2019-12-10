setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force
}

@test "Add apt keys using apt_key ansible module in a different ways" {
	werf build -s :local
}
