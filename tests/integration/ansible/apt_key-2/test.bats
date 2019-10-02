setup() {
	cd tests/integration/ansible/apt_key-2
}

teardown() {
	werf stages purge -s :local
}

@test "Add apt keys using apt_key ansible module in a different ways" {
	werf build -s :local
}
