# buf-plugin-sensitive-guard

Buf check plugin (bufplugin-go) with lint rules for sensitive-field reachability and idempotency keys.

## Rules

- One package, `main`, at the module root, so `go install` names the binary `buf-plugin-sensitive-guard`.
- Rules read everything project-specific from plugin options. No company, product or schema names in code, test data, docs or commits.
- Test with `checktest` and synthetic protos under `testdata/`. Check `buf:lint:ignore` and `--path` behavior with the real Buf CLI, since `checktest` does not apply ignores.
- Protos named in `FilePaths` sit directly under their `testdata/` directory; only imported protos go in subdirectories. On Windows, `checktest` (bufplugin-go v0.11.1) treats a `FilePaths` entry with a directory separator as an import, so rules report nothing.
- Tests: one behavior per test, names `TestUnit_Condition_Outcome`, literal expected values.
- Code and comments are in English. Comments are minimal.
