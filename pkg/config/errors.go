package config

import (
	"bytes"
	"fmt"

	"gopkg.in/flant/yaml.v2"
)

func DumpConfigSection(config interface{}) string {
	res := "```\n"
	d, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}
	res += string(d)
	res += "```\n"

	return res
}

func DumpConfigDoc(doc *Doc) string {
	contentLines := bytes.Split(doc.Content, []byte("\n"))
	if string(contentLines[len(contentLines)-1]) == "" {
		contentLines = contentLines[0 : len(contentLines)-1]
	}

	res := fmt.Sprintf("%s\n\n", doc.RenderFilePath)
	res += "```\n"
	for lineNum, lineBytes := range contentLines {
		res += fmt.Sprintf("%6d  %s\n", doc.Line+lineNum+1, string(lineBytes))
	}
	res += "```\n"

	return res
}
