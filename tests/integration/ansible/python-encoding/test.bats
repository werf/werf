setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --dir app --force
	rm -rf app
}

@test "Run python script with UTF-8 chars inside ansible" {
	git clone repo app
	werf build -s :local --dir app
}
