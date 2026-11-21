package main

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func orderFiles() *checktest.ProtoFileSpec {
	return &checktest.ProtoFileSpec{DirPaths: []string{"testdata/idempotency"}, FilePaths: []string{"order.proto"}}
}

func TestIdempotency_DefaultOptions_ReportsMethodsWithoutKey(t *testing.T) {
	checktest.CheckTest{
		Request: &checktest.RequestSpec{Files: orderFiles(), RuleIDs: []string{idempotencyRuleID}},
		Spec:    spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.CancelOrder does not start with a read verb [Get List], so acme.order.v1.Empty needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 8, 2, 41)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.ShipOrder does not start with a read verb [Get List], so acme.order.v1.OneofKeyRequest needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 9, 2, 49)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.HoldOrder does not start with a read verb [Get List], so acme.order.v1.OptionalKeyRequest needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 10, 2, 52)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.TagOrder does not start with a read verb [Get List], so acme.order.v1.RepeatedKeyRequest needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 11, 2, 51)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.NoteOrder does not start with a read verb [Get List], so acme.order.v1.BytesKeyRequest needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 12, 2, 49)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.Getaway does not start with a read verb [Get List], so acme.order.v1.Empty needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 13, 2, 37)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.DetermineRoute does not start with a read verb [Get List], so acme.order.v1.Empty needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 14, 2, 44)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.RetryOrder does not start with a read verb [Get List], so acme.order.v1.RequestIDRequest needs a singular string field "idempotency_key" outside any oneof.`, FileLocation: at("order.proto", 15, 2, 51)},
		},
	}.Run(t)
}

func TestIdempotency_CustomVerbsAndKey_ReportsAgainstThem(t *testing.T) {
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
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.CreateOrder does not start with a read verb [Get List Determine], so acme.order.v1.KeyedRequest needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 7, 2, 48)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.CancelOrder does not start with a read verb [Get List Determine], so acme.order.v1.Empty needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 8, 2, 41)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.ShipOrder does not start with a read verb [Get List Determine], so acme.order.v1.OneofKeyRequest needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 9, 2, 49)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.HoldOrder does not start with a read verb [Get List Determine], so acme.order.v1.OptionalKeyRequest needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 10, 2, 52)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.TagOrder does not start with a read verb [Get List Determine], so acme.order.v1.RepeatedKeyRequest needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 11, 2, 51)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.NoteOrder does not start with a read verb [Get List Determine], so acme.order.v1.BytesKeyRequest needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 12, 2, 49)},
			{RuleID: idempotencyRuleID, Message: `acme.order.v1.OrderService.Getaway does not start with a read verb [Get List Determine], so acme.order.v1.Empty needs a singular string field "request_id" outside any oneof.`, FileLocation: at("order.proto", 13, 2, 37)},
		},
	}.Run(t)
}
