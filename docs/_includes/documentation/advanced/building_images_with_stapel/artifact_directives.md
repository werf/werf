```yaml
artifact: <artifact_name>
from: <image>
fromLatest: <bool>
herebyIAdmitThatFromLatestMightBreakReproducibility: <bool>
fromCacheVersion: <version>
fromImage: <image_name>
fromArtifact: <artifact_name>
git:
# local git
- add: <absolute path in git repository>
  to: <absolute path inside image>
  owner: <owner>
  group: <group>
  includePaths:
  - <path or glob relative to path in add>
  excludePaths:
  - <path or glob relative to path in add>
  stageDependencies:
    install:
    - <path or glob relative to path in add>
    beforeSetup:
    - <path or glob relative to path in add>
    setup:
    - <path or glob relative to path in add>
# remote git
- url: <git repo url>
  branch: <branch name>
  herebyIAdmitThatBranchMightBreakReproducibility: <bool>
  commit: <commit>
  tag: <tag>
  add: <absolute path in git repository>
  to: <absolute path inside image>
  owner: <owner>
  group: <group>
  includePaths:
  - <path or glob relative to path in add>
  excludePaths:
  - <path or glob relative to path in add>
  stageDependencies:
    install:
    - <path or glob relative to path in add>
    beforeSetup:
    - <path or glob relative to path in add>
    setup:
    - <path or glob relative to path in add>
shell:
  beforeInstall:
  - <cmd>
  install:
  - <cmd>
  beforeSetup:
  - <cmd>
  setup:
  - <cmd>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
ansible:
  beforeInstall:
  - <task>
  install:
  - <task>
  beforeSetup:
  - <task>
  setup:
  - <task>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
mount:
- from: build_dir
  to: <absolute_path>
- from: tmp_dir
  to: <absolute_path>
- fromPath: <absolute_or_relative_path>
  to: <absolute_path>
import:
- artifact: <artifact name>
  image: <image name>
  stage: <stage name>
  before: <install || setup>
  after: <install || setup>
  add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
asLayers: <bool>
```
