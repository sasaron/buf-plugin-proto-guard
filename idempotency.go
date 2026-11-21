package main

import (
	"context"
	"slices"
	"unicode"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	idempotencyRuleID = "IDEMPOTENCY_KEY_REQUIRED"

	idempotencyReadVerbsKey = "idempotency_read_verbs"
	idempotencyKeyFieldKey  = "idempotency_key_field"
)

var (
	defaultReadVerbs = []string{"Get", "List"}
	defaultKeyField  = "idempotency_key"
	idempotencyRule  = &check.RuleSpec{
		ID:      idempotencyRuleID,
		Purpose: "Checks that every RPC whose name does not start with a read verb takes a string idempotency key outside any oneof.",
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewMethodRuleHandler(checkIdempotency, checkutil.WithoutImports()),
	}
)

func checkIdempotency(_ context.Context, w check.ResponseWriter, req check.Request, method protoreflect.MethodDescriptor) error {
	verbs, err := option.GetStringSliceValue(req.Options(), idempotencyReadVerbsKey)
	if err != nil {
		return err
	}
	if len(verbs) == 0 {
		verbs = defaultReadVerbs
	}
	key, err := option.GetStringValue(req.Options(), idempotencyKeyFieldKey)
	if err != nil {
		return err
	}
	if key == "" {
		key = defaultKeyField
	}
	if slices.Contains(verbs, firstWord(string(method.Name()))) || hasKeyField(method.Input(), key) {
		return nil
	}
	w.AddAnnotation(
		check.WithMessagef("%s does not start with a read verb %v, so %s needs a singular string field %q outside any oneof.",
			method.FullName(), verbs, method.Input().FullName(), key),
		check.WithDescriptor(method),
	)
	return nil
}

func hasKeyField(m protoreflect.MessageDescriptor, name string) bool {
	f := m.Fields().ByName(protoreflect.Name(name))
	return f != nil && f.Kind() == protoreflect.StringKind && f.Cardinality() != protoreflect.Repeated && f.ContainingOneof() == nil
}

func firstWord(name string) string {
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			return name[:i]
		}
	}
	return name
}
