package matrix_tests

import (
	"fmt"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/deploy"
)

func RunMatrixTests(tmpDir, projectDir string, werfConfig *config.WerfConfig, opts deploy.RenderOptions) error {
	f, err := LoadConfiguration(projectDir+"/.helm/"+ValuesConfigFilename, "", tmpDir)
	if err != nil {
		return fmt.Errorf("configuration loading error: %v", err)
	}
	defer f.Close()

	f.FindAll()
	err = f.SaveValues()
	if err != nil {
		return fmt.Errorf("saving values error: %v", err)
	}

	c, err := NewModuleController(f.TmpDir, projectDir, werfConfig, opts)
	if err != nil {
		return err
	}
	err = c.Run()
	if err != nil {
		return err
	}
	return nil
}
