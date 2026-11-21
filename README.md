# buf-plugin-sensitive-guard

A [Buf](https://buf.build) check plugin that keeps sensitive fields out of chosen messages and RPCs.

- `SENSITIVE_FIELD_UNREACHABLE` follows message fields recursively from the roots you choose and reports every path to a field marked with your own field option (for example `(acme.option.v1.sensitive) = true`).
- `IDEMPOTENCY_KEY_REQUIRED` is a second, unrelated rule shipped in the same binary: it requires a string idempotency key on every RPC whose name does not start with a read verb.

Tested with Buf CLI v1.75.0 and bufplugin-go v0.11.1.

## Install

```sh
go install github.com/sasaron/buf-plugin-sensitive-guard@latest
```

The binary must be on `$PATH` or referenced by path in `buf.yaml`.

## Configure

Neither rule runs unless listed in `lint.use`.

```yaml
version: v2
lint:
  use:
    - STANDARD
    - SENSITIVE_FIELD_UNREACHABLE
    - IDEMPOTENCY_KEY_REQUIRED
plugins:
  - plugin: buf-plugin-sensitive-guard
    options:
      sensitive_option: acme.option.v1.sensitive
      sensitive_root_message_suffixes: [JobPayload]
      sensitive_root_service_packages: [acme.device.v1]
      sensitive_allowed_external_types:
        - google.protobuf.Timestamp
        - google.protobuf.Duration
      idempotency_read_verbs: [Get, List]
      idempotency_key_field: idempotency_key
```

## SENSITIVE_FIELD_UNREACHABLE

| Option | Meaning |
| --- | --- |
| `sensitive_option` | Required. Full name of a `bool` extension of `google.protobuf.FieldOptions`. Fields set to `true` are sensitive. The plugin reads it by name, so it needs no generated code for your option. |
| `sensitive_root_message_suffixes` | Messages whose name ends with one of these are roots. Nested messages count. |
| `sensitive_root_service_packages` | Request and response types of every RPC in services of these packages are roots, even when the types live in another package. |
| `sensitive_allowed_external_types` | External message types that may be reached. |

At least one kind of root is required.

From each root, the rule follows message fields, map values included, and reports:

- every path to a sensitive field;
- every path to a message type defined outside the files being linted (an import, such as a well-known type or a dependency), unless the type is allowed. The rule cannot see whether such a type carries sensitive data.

Cycles are followed once. A field set to `false` is not sensitive.

Annotations for message roots go on the root's first field of the path. Annotations for service roots go on the RPC. Silence one with a comment above that line:

```protobuf
message AuditJobPayload {
  // buf:lint:ignore SENSITIVE_FIELD_UNREACHABLE the audit job keeps the address.
  Address address = 1;
}
```

`buf lint --path` lints only part of a module. Files outside the path are then imports, so their types are reported as external. Run this rule over whole modules.

## IDEMPOTENCY_KEY_REQUIRED

| Option | Default | Meaning |
| --- | --- | --- |
| `idempotency_read_verbs` | `[Get, List]` | An RPC whose first PascalCase word is one of these is read-only. |
| `idempotency_key_field` | `idempotency_key` | Name of the key field. |

Every other RPC's request must have a singular `string` field with that name outside any oneof. A proto3 `optional` field is in a synthetic oneof, so it does not count. The verb is the whole first word: `Getaway` does not start with `Get`.

Annotations go on the RPC. Silence one with `// buf:lint:ignore IDEMPOTENCY_KEY_REQUIRED <reason>` above it.

## License

Apache-2.0
