setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --dir app --force
	rm -rf app
}

@test "Check exit code of a program that segfaults (https://github.com/werf/werf/issues/1807)" {
	# NOTE: Issues:
	# NOTE:  - https://github.com/ansible/ansible/issues/63123
	# NOTE:  - https://github.com/werf/werf/issues/1807
	# NOTE:
	# NOTE: The bug is in the ansible itself.
	# NOTE: Test will fail when ansible fixes the bug.
	# NOTE: Then test should be reverted to see expected output.
	# NOTE: For now it is OK to see corrupted output of ansible so we test it.

	git clone repo app
	run werf build -s :local --dir app
	[[ "$output" =~ "exit code: -11" ]]
	[[ ! "$output" =~ "Segmentation fault (core dumped)" ]]
}
