package chart_extender

import "text/template"

func SetupIncludeWrapperFuncs(funcMap template.FuncMap) {
	helmIncludeFunc := funcMap["include"].(func(name string, data interface{}) (string, error))
	setupIncludeWrapperFunc := func(name string) {
		funcMap[name] = func(data interface{}) (string, error) {
			return helmIncludeFunc(name, data)
		}
	}

	for _, name := range []string{"werf_image"} {
		setupIncludeWrapperFunc(name)
	}
}
