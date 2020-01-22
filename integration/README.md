# Integration tests

## How to write a test suite

1. Create a directory for the new test suite under `integration` directory in the root of the project:

    ```
    mkdir integration/mytest
    ```

    `mytest` — is the new test suite directory.

2. Create test setup file `integration/mytest/setup_test.go` with the following content:

    ```
    package mytest_test

    import (
    	"testing"

        "github.com/prashantv/gostub"
   
    	"github.com/onsi/ginkgo"
    	"github.com/onsi/gomega"
    	"github.com/onsi/gomega/gexec"

	    "github.com/flant/werf/pkg/testing/utils"
    )

    func TestSuite(t *testing.T) {
    	gomega.RegisterFailHandler(ginkgo.Fail)
    	ginkgo.RunSpecs(t, "Mytest suite")
    }

    var werfBinPath string
    var stubs = gostub.New()

    var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
    	computedPathToWerf := utils.ProcessWerfBinPath()
    	return []byte(computedPathToWerf)
    }, func(computedPathToWerf []byte) {
    	werfBinPath = string(computedPathToWerf)
    })

    var _ = ginkgo.SynchronizedAfterSuite(func() {}, func() {
    	gexec.CleanupBuildArtifacts()
    })

    var _ = ginkgo.BeforeEach(func() {
    	utils.BeforeEachOverrideWerfProjectName(stubs)
    })

    var _ = ginkgo.AfterEach(func() {
    	stubs.Reset()
    })
    ```

    **NOTE**: Go package should be named with suffix `_test` (i.e. `mytest_test`).

3. The actual tests should reside separately from the setup code in `setup_test.go` file. Create a new file with arbitrary name for the new test case.

    For example, let's test a component named "Config loader" (this name usually goes to the `Describe` ginkgo statement, see example below). All tests for this component will reside in the file `integration/mytest/config_loader_test.go` (note that `_test.go` suffix is required for all go test files).

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

5. Place more files with `_test.go` to test more cases under the same test suite in the directory `integration/mytest`.

## Testing a werf project

### werf project name generation

When running tests with `ginkgo` (or `go test`) werf temporary project name will be generated for each separate `It` statement. This behaviour is defined for the each test suite in `setup_test.go` file.

So you cannot use the same werf project between different `It` blocks.

Furthermore, different `It` blocks (within the same `Describe` or different `Describe` — does not matter) could run in parallel.

### `Describe`, `Context` and `It`

When your test uses werf project it is convinient to create a special `Context` block within `Describe` for each project.

In this `Context` single `It` block should be defined for the actual test code and `AfterEach` block to perform cleanup for the project (typically `werf stages purge --force` and `werf dismiss ...`).

For each new `It` block define a surrounding `Context` block and `AfterEach` cleanup block.

```
var _ = Describe("COMPONENT", func() {
    Context("when ...", func() { // Context for the project test 1
        AfterEach(func() {
            // Run `werf stages purge --force` and/or `werf dismiss ...`
            // Run more cleanup for the project
        })

        It("should ...", func(done Done) {
            // Test code
            // Run `werf build` and/or `werf deploy`
            // Check results
            // ...

            close(done)
        }, 300)
    })

    Context("when ...", func() { // Context for the project test 2
        AfterEach(func() {
            // Run `werf stages purge --force` and/or `werf dismiss ...`
            // Run more cleanup for the project
        })

        It("should ...", func(done Done) {
            // Test code
            // Run `werf build` and/or `werf deploy`
            // Check results
            // ...

            close(done)
        }, 300)
    })

    // ...
})
```

### Catch and test werf output in realtime

`github.com/flant/werf/pkg/testing/utils/liveexec` package provides execution of external commands with realtime output handling and ability to fail fast when expectation of output was not met. Example of liveexec usage:

```
func werfDeploy(dir string, opts liveexec.ExecCommandOptions, extraArgs ...string) error {
	return liveexec.ExecCommand(dir, werfBinPath, opts, append([]string{"deploy", "--env", "dev"}, extraArgs...)...)
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
