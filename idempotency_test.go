package main

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func orderFiles() *checktest.ProtoFileSpec {
	return &checktest.ProtoFileSpec{DirPaths: []string{"testdata/idempotency"}, FilePaths: []string{"order.proto"}}
}

func missingKey(method, input string, verbs, key string, line, endCol int) checktest.ExpectedAnnotation {
	return checktest.ExpectedAnnotation{
		RuleID:       idempotencyRuleID,
		Message:      "acme.order.v1.OrderService." + method + " does not start with a read verb " + verbs + ", so acme.order.v1." + input + " needs a singular string field \"" + key + "\" outside any oneof.",
		FileLocation: at("order.proto", line, 2, line, endCol),
	}
}

func TestIdempotency_DefaultOptions_ReportsMethodsWithoutKey(t *testing.T) {
	const verbs = "[Get List]"
	checktest.CheckTest{
		Request: &checktest.RequestSpec{Files: orderFiles(), RuleIDs: []string{idempotencyRuleID}},
		Spec:    spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			missingKey("CancelOrder", "Empty", verbs, "idempotency_key", 8, 41),
			missingKey("ShipOrder", "OneofKeyRequest", verbs, "idempotency_key", 9, 49),
			missingKey("HoldOrder", "OptionalKeyRequest", verbs, "idempotency_key", 10, 52),
			missingKey("TagOrder", "RepeatedKeyRequest", verbs, "idempotency_key", 11, 51),
			missingKey("NoteOrder", "BytesKeyRequest", verbs, "idempotency_key", 12, 49),
			missingKey("Getaway", "Empty", verbs, "idempotency_key", 13, 37),
			missingKey("DetermineRoute", "Empty", verbs, "idempotency_key", 14, 44),
			missingKey("RetryOrder", "RequestIDRequest", verbs, "idempotency_key", 15, 51),
		},
	}.Run(t)
}

func TestIdempotency_CustomVerbsAndKey_ReportsAgainstThem(t *testing.T) {
	const verbs = "[Get List Determine]"
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files:   orderFiles(),
			RuleIDs: []string{idempotencyRuleID},
			Options: map[string]any{
				idempotencyReadVerbsKey: []string{"Get", "List", "Determine"},
				idempotencyKeyFieldKey:  "request_id",
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			missingKey("CreateOrder", "KeyedRequest", verbs, "request_id", 7, 48),
			missingKey("CancelOrder", "Empty", verbs, "request_id", 8, 41),
			missingKey("ShipOrder", "OneofKeyRequest", verbs, "request_id", 9, 49),
			missingKey("HoldOrder", "OptionalKeyRequest", verbs, "request_id", 10, 52),
			missingKey("TagOrder", "RepeatedKeyRequest", verbs, "request_id", 11, 51),
			missingKey("NoteOrder", "BytesKeyRequest", verbs, "request_id", 12, 49),
			missingKey("Getaway", "Empty", verbs, "request_id", 13, 37),
		},
	}.Run(t)
}
