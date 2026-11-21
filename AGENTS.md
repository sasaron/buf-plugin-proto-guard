# buf-plugin-proto-guard

Buf check plugin (bufplugin-go) with lint rules for sensitive-field reachability and idempotency keys.

## Rules

- One package, `main`, at the module root, so `go install` names the binary `buf-plugin-proto-guard`.
- Rules read everything project-specific from plugin options. No company, product or schema names in code, test data, docs or commits.
- Test with `checktest` and synthetic protos under `testdata/`. Check `buf:lint:ignore` and `--path` behavior with the real Buf CLI, since `checktest` does not apply ignores.
- Tests: one behavior per test, names `TestUnit_Condition_Outcome`, literal expected values.
- Code and comments are in English. Comments are minimal.
