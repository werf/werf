package config

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/werf/werf/pkg/util"
)

type configError struct {
	s string
}

func (e *configError) Error() string {
	return e.s
}

func newConfigError(message string) error {
	return &configError{message}
}

func newDetailedConfigError(message string, configSection interface{}, configDoc *doc) error {
	var errorString string
	if configSection != nil {
		errorString = fmt.Sprintf("%s\n\n%s\n%s", message, dumpConfigSection(configSection), dumpConfigDoc(configDoc))
	} else {
		errorString = fmt.Sprintf("%s\n\n%s", message, dumpConfigDoc(configDoc))
	}
	return newConfigError(errorString)
}

func getLines(data []byte) [][]byte {
	contentLines := bytes.Split(data, []byte("\n"))
	if string(contentLines[len(contentLines)-1]) == "" {
		contentLines = contentLines[0 : len(contentLines)-1]
	}
	return contentLines
}

func dumpConfigSection(config interface{}) string {
	d, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}

	contentLines := getLines(d)

	res := ""
	for _, lineBytes := range contentLines {
		res += fmt.Sprintf("    %s\n", string(lineBytes))
	}

	return res
}

func dumpConfigDoc(doc *doc) string {
	res := fmt.Sprintf("%s\n\n", doc.RenderFilePath)
	res += util.NumerateLines(string(doc.Content), doc.Line+1)
	res += "\n"

	return res
}
