setup() {
	cd tests/integration/ansible/yum
}

teardown() {
	werf stages purge -s :local
}

@test "Ansible successfully install packages into centos image using yum module (FIXME https://github.com/flant/werf/issues/1798)" {
	skip
	werf build -s :local
}
