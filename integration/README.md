# Integration tests

To create an integration test suite you need:

 1. Write a golang test suite itself.
 2. Enable your suite for CI.

## How to write a test suite

1. Create a directory for the new test suite under `integration/suites` directory in the root of the project:

    ```
    mkdir integration/suites/mytest
    ```

    `mytest` — is the new test suite directory.

    **NOTE**: Suites directories can be grouped and arbitrary nested, for example: `integration/suites/my_suites_group/another_subgroup/mysuite`.

2. Create suite setup file `integration/suites/mytest/suite_test.go` using following content example:

    ```
    package ansible_test

    import (
      "testing"
    
      "github.com/werf/werf/test/pkg/suite_init"
    )
    
    var testSuiteEntrypointFunc = suite_init.MakeTestSuiteEntrypointFunc("MYTEST suite", suite_init.TestSuiteEntrypointFuncOptions{})
    
    func TestSuite(t *testing.T) {
        testSuiteEntrypointFunc(t)
    }
    
    var SuiteData suite_init.SuiteData
    
    var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
    var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
    var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
    var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
    ```

    **NOTE**: Go package should be named with suffix `_test` (i.e. `mytest_test`).

3. The actual tests should reside separately from the setup code in `suite_test.go` file. Create a new file with arbitrary name for the new test case.

    For example, let's test a component named "Config loader" (this name usually goes to the `Describe` ginkgo statement, see example below). All tests for this component will reside in the file `integration/suites/mytest/config_loader_test.go` (note that `_test.go` suffix is required for all go test files).

4. Use `Describe`, `Context`, `It` and other ginkgo statements to organize tests inside your file. For example:

    ```
    var _ = Describe("Config loader", func() {
    	Context("when using correct config", func() {
    		AfterEach(func() {
                // Cleanup code
    		})

    		It("should load config without errors", func(done Done) {
    			// Test code

    			close(done)
    		}, 300)
    	})

        Context("when using config with errors", func() {
    		AfterEach(func() {
                // Cleanup code
    		})

    		It("should fail to load config", func(done Done) {
    			// Test code

    			close(done)
    		}, 300)
    	})

        // ...
    })
    ```

5. Place more files with `_test.go` to test more cases under the same test suite in the directory `integration/suites/mytest`.

## Testing a werf project

### Suite data initialization

There is `integration/pkg/suite_init` package which provides objects and helpers to initialize some typically needed data for your suite. This package for example provides:
 - container registry per implementation data;
 - project name data;
 - tmp dir data;
 - etc.

To setup needed data you should create following `SuiteData` variable in the `suite_test.go` setup file:

```
import "github.com/werf/werf/test/pkg/suite_init"

var SuiteData suite_init.SuiteData
```

Then you should call corresponding setup function to setup needed data, providing required arguments. Note that suite data can depend on each other and should be initialized in the corresponding order. For example, let's initialize stubs, project name, synchronized suite callbacks and werf binary:

```
import "github.com/werf/werf/test/pkg/suite_init"

var SuiteData suite_init.SuiteData

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupSynchronizedSuiteCallbacks(suite_init.NewSynchronizedSuiteCallbacksData())
var _ = SuiteData.SetupWerfBinary(suite_init.NewWerfBinaryData(SuiteData.SynchronizedSuiteCallbacksData))
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
```

### Note on werf project name generation

When running tests with `ginkgo` (or `go test`) werf temporary project name will be generated for each separate `It` statement. This behaviour is defined for the each test suite in `setup_test.go` file.

So you cannot use the same werf project between different `It` blocks.

Furthermore, different `It` blocks (within the same `Describe` or different `Describe` — does not matter) could run in parallel.

Use `integration/pkg/suite_init/project_name_data.go` suite data initialization to generate project name:

```
import "github.com/werf/werf/test/pkg/suite_init"

var SuiteData suite_init.SuiteData

var _ = SuiteData.SetupStubs(suite_init.NewStubsData())
var _ = SuiteData.SetupProjectName(suite_init.NewProjectNameData(SuiteData.StubsData))
```

### `Describe`, `Context` and `It`

When your test uses werf project it is convenient to create a special `Context` block within `Describe` for each project.

In this `Context` single `It` block should be defined for the actual test code and `AfterEach` block to perform cleanup for the project (typically `werf purge --force` and `werf dismiss ...`).

For each new `It` block define a surrounding `Context` block and `AfterEach` cleanup block.

```
var _ = Describe("COMPONENT", func() {
    Context("when ...", func() { // Context for the project test 1
        AfterEach(func() {
            // Run `werf purge --force` and/or `werf dismiss ...`
            // Run more cleanup for the project
        })

        It("should ...", func(done Done) {
            // Test code
            // Run `werf build` and/or `werf converge`
            // Check results
            // ...

            close(done)
        }, 300)
    })

    Context("when ...", func() { // Context for the project test 2
        AfterEach(func() {
            // Run `werf purge --force` and/or `werf dismiss ...`
            // Run more cleanup for the project
        })

        It("should ...", func(done Done) {
            // Test code
            // Run `werf build` and/or `werf converge`
            // Check results
            // ...

            close(done)
        }, 300)
    })

    // ...
})
```

### Catch and test werf output in realtime

`github.com/werf/werf/test/pkg/utils/liveexec` package provides execution of external commands with realtime output handling and ability to fail fast when expectation of output was not met. Example of liveexec usage:

```
func werfDeploy(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, append([]string{"deploy"}, extraArgs...)...)
}

It("Should ...", func(done Done) {
    gotDeletingHookLine := false

    Expect(werfDeploy("helm_hooks_deleter_app1", liveexec.ExecCommandOptions{
        OutputLineHandler: func(line string) {
            Expect(strings.HasPrefix(line, "│ NOTICE Will not delete Job/migrate: resource does not belong to the helm release")).ShouldNot(BeTrue(), fmt.Sprintf("Got unexpected output line: %v", line))

            if strings.HasPrefix(line, "│ Deleting resource Job/migrate from release") {
                gotDeletingHookLine = true
            }
        },
    })).Should(Succeed())

    Expect(gotDeletingHookLine).To(BeTrue())

    close(done)
}, 300)
```

## Enable your suite for CI

Werf's CI provides following layouts to run integration tests:

 - **default**. Default docker registry, all supported OS (linux, windows, macos), NO k8s available. Use this layout as a default choise when you don't need an k8s cluster. This layout could use either persistent or ephemeral runners.
 - **k8s_per_version**. K8s cluster (multiple versions of kubernetes itself are tested), default docker registry accessible from k8s cluster, all supported OS (linux, windows, macos). Use this layout when you need an k8s cluster. This layout could use either persistent or ephemeral runners.
 - **container_registry_per_implementation**. Multiple implementations of container registries are available (github packages, harbor, gcr, aws ecr, azure, etc.), all supported OS (linux, windows, macos), NO k8s available. Use this layout when you need to test against all known container registries implementations.
 - **k8s_per_version_and_container_registry_per_implementation**. Multiple implementations of container registries are available (github packages, harbor, gcr, aws ecr, azure, etc.), all supported OS (linux, windows, macos), k8s cluster (multiple versions of kubernetes itself are tested). This is the most advanced layout which should be used when you need k8s and all container registries implementations available.  

To enable test in some layout you need to create symlink in `integration/ci_suites/LAYOUT/MYSUITE` to your suite which reside in the common suites dir `integration/suites/MYSUITE`:

```
integration/ci_suites/
├── container_registry_per_implementation
│   └── cleanup -> ../../suites/cleanup/
├── default
│   ├── ansible -> ../../suites/ansible
│   ├── build -> ../../suites/build
│   ├── ci-env -> ../../suites/ci-env
│   ├── config -> ../../suites/config
│   ├── helm
│   │   ├── get_something -> ../../../suites/helm/get_something
│   │   ├── render -> ../../../suites/helm/render/
│   │   └── secret -> ../../../suites/helm/secret/
│   └── managed_images -> ../../suites/managed_images
├── k8s_per_version
│   ├── cleanup_after_converge -> ../../suites/cleanup_after_converge/
│   ├── helm
│   │   └── deploy_rollback -> ../../../suites/helm/deploy_rollback/
│   └── releaseserver -> ../../suites/releaseserver/
└── k8s_per_version_and_container_registry_per_implementation
    └── bundles -> ../../suites/bundles/
```
