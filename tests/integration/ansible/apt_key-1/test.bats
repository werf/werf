setup() {
	cd tests/integration/ansible/apt_key-1
}

teardown() {
	werf stages purge -s :local --force --dir project-1
	werf stages purge -s :local --force --dir project-2
}

@test "Ansible fails to install package without a key and succeeds with the key" {
	run werf build -s :local --dir project-1
	[ "$status" -eq 1 ]
	[[ "$output" =~ "public key is not available: NO_PUBKEY" ]]

	run werf build -s :local --dir project-2
	[ "$status" -eq 0 ]
	[[ "$output" =~ "apt 'Install package from new repository' [clickhouse-client] (".*" seconds)" ]]
	[[ ! "$output" =~ "apt 'Install package from new repository' [clickhouse-client] (".*" seconds) FAILED" ]]
}
