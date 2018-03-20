package ruby2go

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flant/dapp/pkg/image"
)

var (
	WorkingDir       string
	ArgsFromFilePath string
	ResultToFilePath string
)

func usage(progname string) {
	fmt.Fprintf(os.Stderr, "%s\n", progname)
	flag.PrintDefaults()
	os.Exit(2)
}

func readJsonObjectFromFile(path string) (map[string]interface{}, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("No such file %s (%s)\n", path, err)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func writeJsonObjectToFile(obj map[string]interface{}, path string) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func RunCli(progname string, runFunc func(map[string]interface{}) (map[string]interface{}, error)) {
	WorkingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine working dir: %s\n", err)
		os.Exit(1)
	}
	_ = WorkingDir

	flag.Usage = func() { usage(progname) }
	flag.StringVar(&ArgsFromFilePath, "args-from-file", "", "path to json file with input parameters")
	flag.StringVar(&ResultToFilePath, "result-to-file", "", "path to json file with program output")
	flag.Parse()

	if ArgsFromFilePath == "" {
		fmt.Fprintf(os.Stderr, "`-args-from-file` param required!\n")
		os.Exit(1)
	}
	if ResultToFilePath == "" {
		fmt.Fprintf(os.Stderr, "`-result-to-file` param required!\n")
		os.Exit(1)
	}

	argsMap, err := readJsonObjectFromFile(ArgsFromFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read args json object from file %s: %s\n", ArgsFromFilePath, err)
		os.Exit(1)
	}

	exitCode := 0
	resultMap := make(map[string]interface{})

	resultMap["data"], err = runFunc(argsMap)
	if err != nil {
		resultMap["error"] = fmt.Sprintf("%s", err)
		exitCode = 16
	}

	err = writeJsonObjectToFile(resultMap, ResultToFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write result json object to file %s: %s", ResultToFilePath, err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func CommandWithImage(args map[string]interface{}, command func(stageImage *image.Stage) error) (map[string]interface{}, error) {
	stageImage, err := stageImageFromArgs(args)
	if err != nil {
		return nil, err
	}

	if err := command(stageImage); err != nil {
		return nil, err
	}

	resultMap, err := stageImageToArgs(stageImage, make(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func stageImageFromArgs(args map[string]interface{}) (*image.Stage, error) {
	rubyImage := &Stage{}

	switch args["image"].(type) {
	case string:
		if err := json.Unmarshal([]byte(args["image"].(string)), rubyImage); err != nil {
			return nil, fmt.Errorf("image field unmarshaling failed: %s", err.Error())
		}
		return rubyStageToImageStage(rubyImage), nil
	default:
		return nil, fmt.Errorf("image field value `%v` isn't supported", args["image"])
	}
}

func stageImageToArgs(stageImage *image.Stage, args map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(imageStageToRubyStage(stageImage))
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("stage marshaling failed: %s", err.Error()))
	}
	args["image"] = string(raw)
	return args, nil
}
