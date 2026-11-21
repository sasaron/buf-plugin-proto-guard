// Command buf-plugin-sensitive-guard is a Buf check plugin that keeps sensitive
// fields out of chosen messages and RPCs. It also has a rule that requires
// idempotency keys on state-changing RPCs.
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
		LicenseURL:    "https://github.com/sasaron/buf-plugin-sensitive-guard/blob/main/LICENSE",
	},
}

func main() {
	check.Main(spec)
}
