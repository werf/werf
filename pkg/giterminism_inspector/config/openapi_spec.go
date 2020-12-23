package config

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/go-openapi/validate/post"
	"github.com/hashicorp/go-multierror"
	"sigs.k8s.io/yaml"
)

const (
	// TODO: use embedded file openapi_spec.yaml instead
	schemaYaml = `type: object
required:
- giterminismConfigVersion
additionalProperties: {}
properties:
  giterminismConfigVersion:
    type: string
    enum: ["1"]
  config:
    $ref: '#/definitions/Config'
definitions:
  Config:
    type: object
    additionalProperties: {}
    properties:
      allowUncommitted:
        type: boolean
`
)

func openAPISchema() *spec.Schema {
	hash := map[string]interface{}{}
	if err := yaml.UnmarshalStrict([]byte(schemaYaml), &hash); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	data, err := json.Marshal(hash)
	if err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	schema := &spec.Schema{}
	if err := json.Unmarshal(data, schema); err != nil {
		panic(fmt.Sprint("unexpected error: ", err))
	}

	err = spec.ExpandSchema(schema, schema, nil)
	if err != nil {
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

	var allErrs *multierror.Error
	allErrs = multierror.Append(allErrs, result.Errors...)

	return allErrs.ErrorOrNil()
}
