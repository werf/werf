setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --dir app --force
	rm -rf app
}

@test "Run yarn module to install nodejs packages" {
	git clone repo app
	werf build -s :local --dir app
}
