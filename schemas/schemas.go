// Package schemas holds the JSON Schemas of werf configuration files.
//
// The files are published for IDEs and validated by the unit tests of the
// packages that parse the corresponding configuration; the schemas werf
// itself validates against at runtime are embedded here.
package schemas

import _ "embed"

//go:embed werf-giterminism.json
var GiterminismConfig string
