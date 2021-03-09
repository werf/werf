package helpers

import (
	"context"
	"text/template"

	"github.com/werf/logboek"
)

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

func SetupWerfImageDeprecationFunc(ctx context.Context, funcMap template.FuncMap) {
	funcMap["_print_werf_image_deprecation"] = func() (string, error) {
		logboek.Context(ctx).Warn().LogF("DEPRECATION WARNING: Usage of werf_image is deprecated, use .Values.werf.image.IMAGE_NAME directly instead, werf_image template function will be removed in v1.3.\n")
		return "", nil
	}
}
