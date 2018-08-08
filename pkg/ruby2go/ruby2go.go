package ruby2go

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flant/dapp/pkg/util"
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

func RunCli(progname string, runFunc func(map[string]interface{}) (interface{}, error)) {
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

func CommandFieldFromArgs(args map[string]interface{}) (string, error) {
	return StringFieldFromMapInterface("command", args)
}

func OptionsFieldFromArgs(args map[string]interface{}) (map[string]interface{}, error) {
	value, ok := args["options"]
	if ok {
		return util.InterfaceToMapStringInterface(value)
	} else {
		return nil, fmt.Errorf("options field value `%#v` can't be casting into map[string]interface{}", args["options"])
	}
}

func StringOptionFromArgs(option string, args map[string]interface{}) (string, error) {
	options, err := OptionsFieldFromArgs(args)
	if err != nil {
		return "", err
	}
	return StringFieldFromMapInterface(option, options)
}

func StringFieldFromMapInterface(field string, value map[string]interface{}) (string, error) {
	switch value[field].(type) {
	case string:
		return value[field].(string), nil
	default:
		return "", fmt.Errorf("option `%s` field value `%#v` can't be casting into string", field, value[field])
	}
}
