// Package schema embeds the normative HCR JSON Schema document so it ships
// inside the jigctl binary instead of being read from disk at runtime.
package schema

import _ "embed"

//go:embed hcr.schema.json
var HCR []byte
