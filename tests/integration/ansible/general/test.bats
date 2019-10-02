setup() {
	cd tests/integration/ansible/general
}

teardown() {
	werf stages purge -s :local
}

@test "Ansible modules should perform without errors" {
	werf build -s :local
}
