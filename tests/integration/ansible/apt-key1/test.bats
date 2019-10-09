setup() {
	cd $BATS_TEST_DIRNAME
}

teardown() {
	werf stages purge -s :local --force --dir app1
	werf stages purge -s :local --force --dir app2
}

@test "Ansible fails to install package without a key and succeeds with the key" {
	run werf build -s :local --dir app1
	[ "$status" -eq 1 ]
	[[ "$output" =~ "public key is not available: NO_PUBKEY" ]]

	run werf build -s :local --dir app2
	[ "$status" -eq 0 ]
	[[ "$output" =~ "apt 'Install package from new repository' [clickhouse-client] (".*" seconds)" ]]
	[[ ! "$output" =~ "apt 'Install package from new repository' [clickhouse-client] (".*" seconds) FAILED" ]]
}
