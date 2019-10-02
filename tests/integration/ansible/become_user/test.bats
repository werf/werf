setup() {
	cd tests/integration/ansible/become_user
}

teardown() {
	werf stages purge -s :local
}

@test "become_user module perform without errors (FIXME https://github.com/flant/werf/issues/1806)" {
	skip
	werf build -s :local
}
