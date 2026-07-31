//go:build embedwerfdeno

package deno

import _ "embed"

//go:embed embed/linux/arm64/deno.gz
var embeddedDeno []byte
