package config

import (
	"bytes"
	"fmt"

	"gopkg.in/flant/yaml.v2"
)

type ConfigError struct {
	s string
}

func (e *ConfigError) Error() string {
	return e.s
}

func NewConfigError(message string) error {
	return &ConfigError{message}
}

func NewDetailedConfigError(message string, configSection interface{}, configDoc *Doc) error {
	var errorString string
	if configSection != nil {
		errorString = fmt.Sprintf("%s\n\n%s\n%s", message, DumpConfigSection(configSection), DumpConfigDoc(configDoc))
	} else {
		errorString = fmt.Sprintf("%s\n\n%s", message, DumpConfigDoc(configDoc))
	}
	return NewConfigError(errorString)
}

func getLines(data []byte) [][]byte {
	contentLines := bytes.Split(data, []byte("\n"))
	if string(contentLines[len(contentLines)-1]) == "" {
		contentLines = contentLines[0 : len(contentLines)-1]
	}
	return contentLines
}

func DumpConfigSection(config interface{}) string {
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

func DumpConfigDoc(doc *Doc) string {
	contentLines := getLines(doc.Content)

	res := fmt.Sprintf("%s\n\n", doc.RenderFilePath)
	for lineNum, lineBytes := range contentLines {
		res += fmt.Sprintf("%6d  %s\n", doc.Line+lineNum+1, string(lineBytes))
	}
	res += "\n"

	return res
}
