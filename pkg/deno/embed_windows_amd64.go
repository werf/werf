//go:build embedwerfdeno

package deno

import _ "embed"

//go:embed embed/windows/amd64/deno.gz
var embeddedDeno []byte
