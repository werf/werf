package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/go-openapi/validate/post"
	"sigs.k8s.io/yaml"

	"github.com/werf/werf/v2/schemas"
)

func openAPISchema() *spec.Schema {
	schema := &spec.Schema{}
	if err := json.Unmarshal([]byte(schemas.GiterminismConfig), schema); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	if err := spec.ExpandSchema(schema, schema, nil); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	return schema
}

func processWithOpenAPISchema(dataObj *[]byte) error {
	validator := validate.NewSchemaValidator(openAPISchema(), nil, "", strfmt.Default)

	var blank map[string]interface{}
	err := yaml.Unmarshal(*dataObj, &blank)
	if err != nil {
		return err
	}

	result := validator.Validate(blank)
	if result.IsValid() {
		post.ApplyDefaults(result)
		*dataObj, err = json.Marshal(result.Data())
		if err != nil {
			panic(fmt.Sprint("unexpected error: ", err))
		}

		return nil
	}

	return errors.Join(result.Errors...)
}
