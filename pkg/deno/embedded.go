//go:build embedwerfdeno

package deno

// EmbeddedDenoData returns the compressed embedded Deno binary. Whether it is usable is nelm's call,
// not werf's: it checks the bytes against the release its lock pins, which is the same lock the
// generator that produced them verified against. There is a blob only for the platforms that lock
// pins, so a tagged build for any other one fails with "undefined: embeddedDeno".
func EmbeddedDenoData() ([]byte, bool) {
	return embeddedDeno, true
}
