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
	schemaYaml = `type: object
required:
  - giterminismConfigVersion
additionalProperties: {}
properties:
  giterminismConfigVersion:
    oneOf:
      - type: number
        enum: [1]
      - type: string
        enum: ["1"]
  cli:
    $ref: '#/definitions/CLI'
  config:
    $ref: '#/definitions/Config'
  helm:
    $ref: '#/definitions/Helm'
definitions:
  CLI:
    type: object
    additionalProperties: {}
    properties:
      allowCustomTags:
        type: boolean
  Config:
    type: object
    additionalProperties: {}
    properties:
      allowUncommitted:
        type: boolean
      allowUncommittedTemplates:
        type: array
        items:
          type: string
      goTemplateRendering:
        $ref: '#/definitions/ConfigGoTemplateRendering'
      stapel:
        $ref: '#/definitions/ConfigStapel'
      dockerfile:
        $ref: '#/definitions/ConfigDockerfile'
  ConfigGoTemplateRendering:
    type: object
    additionalProperties: {}
    properties:
      allowEnvVariables:
        type: array
        items:
          type: string
      allowUncommittedFiles:
        type: array
        items:
          type: string
  ConfigStapel:
    type: object
    additionalProperties: {}
    properties:
      allowFromLatest:
        type: boolean
      git:
        $ref: '#/definitions/ConfigStapelGit'
      mount:
        $ref: '#/definitions/ConfigStapelMount'
  ConfigStapelGit:
    type: object
    additionalProperties: {}
    properties:
      allowBranch:
        type: boolean
  ConfigStapelMount:
    type: object
    additionalProperties: {}
    properties:
      allowBuildDir:
        type: boolean
      allowFromPaths:
        type: array
        items:
          type: string
  ConfigDockerfile:
    type: object
    additionalProperties: {}
    properties:
      allowUncommitted:
        type: array
        items:
          type: string
      allowUncommittedDockerignoreFiles:
        type: array
        items:
          type: string
      allowContextAddFiles:
        type: array
        items:
          type: string
  Helm:
    type: object
    additionalProperties: {}
    properties:
      allowUncommittedFiles:
        type: array
        items:
          type: string
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
