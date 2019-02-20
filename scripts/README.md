# Release process

Main release script:

`scripts/release.sh TAG` performs following steps from local host machine.

1. Create release message and save into git tag (it is required to push git tag into origin)

2. Build and upload release binaries.

3. Publish release by the specified tag:
   - sign uploaded binaries;
   - create github release.
