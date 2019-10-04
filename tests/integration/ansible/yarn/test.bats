setup() {
	cd tests/integration/ansible/yarn
}

teardown() {
	werf stages purge -s :local --dir app
	rm -rf app
}

@test "Run yarn module to install nodejs packages" {
	git clone repo app
	werf build -s :local --dir app
}
