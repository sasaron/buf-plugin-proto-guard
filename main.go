// Command buf-plugin-proto-guard is a Buf check plugin with lint rules that keep
// sensitive fields out of chosen messages and RPCs, and require idempotency keys
// on state-changing RPCs.
package main

import (
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
)

var spec = &check.Spec{
	Rules: []*check.RuleSpec{
		sensitiveRule,
		idempotencyRule,
	},
	Info: &info.Spec{
		Documentation: "Lint rules that keep sensitive fields out of chosen messages and RPCs, and require idempotency keys on state-changing RPCs.",
		SPDXLicenseID: "apache-2.0",
		LicenseURL:    "https://github.com/sasaron/buf-plugin-proto-guard/blob/main/LICENSE",
	},
}

func main() {
	check.Main(spec)
}
