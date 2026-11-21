package main

import (
	"context"
	"strings"
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
)

func sensitiveFiles(paths ...string) *checktest.ProtoFileSpec {
	return &checktest.ProtoFileSpec{DirPaths: []string{"testdata/sensitive"}, FilePaths: paths}
}

func at(file string, line, startCol, endCol int) *checktest.ExpectedFileLocation {
	return &checktest.ExpectedFileLocation{FileName: file, StartLine: line, StartColumn: startCol, EndLine: line, EndColumn: endCol}
}

func TestSpec_Defined_IsValid(t *testing.T) {
	checktest.SpecTest(t, spec)
}

func TestSensitive_SuffixRoots_ReportsFirstFieldOfEachPath(t *testing.T) {
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files:   sensitiveFiles("app.proto"),
			RuleIDs: []string{sensitiveRuleID},
			Options: map[string]any{
				sensitiveOptionKey:          "acme.option.v1.sensitive",
				sensitiveRootSuffixesKey:    []string{"Payload"},
				sensitiveAllowedExternalKey: []string{"google.protobuf.Timestamp"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{RuleID: sensitiveRuleID, Message: "acme.app.v1.ShipPayload reaches acme.app.v1.ShipPayload.customer.address.line.", FileLocation: at("app.proto", 22, 2, 24)},
			{RuleID: sensitiveRuleID, Message: "acme.app.v1.TagPayload reaches acme.app.v1.TagPayload.addresses.value.line.", FileLocation: at("app.proto", 27, 2, 37)},
			{RuleID: sensitiveRuleID, Message: "acme.app.v1.Batch.ItemPayload reaches acme.app.v1.Batch.ItemPayload.note of external type acme.shared.v1.Note.", FileLocation: at("app.proto", 32, 4, 33)},
		},
	}.Run(t)
}

func TestSensitive_ExternalTypeNotAllowed_ReportsExternalType(t *testing.T) {
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files:   sensitiveFiles("app.proto"),
			RuleIDs: []string{sensitiveRuleID},
			Options: map[string]any{
				sensitiveOptionKey:       "acme.option.v1.sensitive",
				sensitiveRootSuffixesKey: []string{"ShipPayload"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{RuleID: sensitiveRuleID, Message: "acme.app.v1.ShipPayload reaches acme.app.v1.ShipPayload.customer.address.line.", FileLocation: at("app.proto", 22, 2, 24)},
			{RuleID: sensitiveRuleID, Message: "acme.app.v1.ShipPayload reaches acme.app.v1.ShipPayload.created_at of external type google.protobuf.Timestamp.", FileLocation: at("app.proto", 23, 2, 43)},
		},
	}.Run(t)
}

func TestSensitive_ServicePackageRoots_ReportsMethod(t *testing.T) {
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files:   sensitiveFiles("api.proto"),
			RuleIDs: []string{sensitiveRuleID},
			Options: map[string]any{
				sensitiveOptionKey:       "acme.option.v1.sensitive",
				sensitiveRootPackagesKey: []string{"acme.api.v1"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{RuleID: sensitiveRuleID, Message: "acme.api.v1.CustomerService.GetCustomer response is external type acme.app.v1.Customer.", FileLocation: at("api.proto", 7, 2, 69)},
		},
	}.Run(t)
}

func TestSensitive_ServicePackageRootsWithTypesLinted_ReportsSensitiveField(t *testing.T) {
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files:   sensitiveFiles("api.proto", "app.proto"),
			RuleIDs: []string{sensitiveRuleID},
			Options: map[string]any{
				sensitiveOptionKey:       "acme.option.v1.sensitive",
				sensitiveRootPackagesKey: []string{"acme.api.v1"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{RuleID: sensitiveRuleID, Message: "acme.api.v1.CustomerService.GetCustomer response reaches acme.app.v1.Customer.address.line.", FileLocation: at("api.proto", 7, 2, 69)},
		},
	}.Run(t)
}

func TestSensitive_InvalidOptions_ReturnsError(t *testing.T) {
	for name, tc := range map[string]struct {
		options map[string]any
		want    string
	}{
		"missing_option": {
			options: map[string]any{sensitiveRootSuffixesKey: []string{"Payload"}},
			want:    `"sensitive_option"`,
		},
		"unknown_extension": {
			options: map[string]any{sensitiveOptionKey: "acme.option.v1.missing", sensitiveRootSuffixesKey: []string{"Payload"}},
			want:    `sensitive_option "acme.option.v1.missing"`,
		},
		"non-bool_extension": {
			options: map[string]any{sensitiveOptionKey: "acme.option.v1.label", sensitiveRootSuffixesKey: []string{"Payload"}},
			want:    `"acme.option.v1.label" is not a bool extension`,
		},
		"no_roots": {
			options: map[string]any{sensitiveOptionKey: "acme.option.v1.sensitive"},
			want:    `"sensitive_root_message_suffixes" or "sensitive_root_service_packages"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := runCheck(t, &checktest.RequestSpec{
				Files:   sensitiveFiles("app.proto"),
				RuleIDs: []string{sensitiveRuleID},
				Options: tc.options,
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Check(options = %v) error = %v, want it to contain %q", tc.options, err, tc.want)
			}
		})
	}
}

func TestSensitive_OtherRuleOnly_IgnoresMissingOption(t *testing.T) {
	err := runCheck(t, &checktest.RequestSpec{
		Files:   sensitiveFiles("app.proto"),
		RuleIDs: []string{idempotencyRuleID},
	})
	if err != nil {
		t.Errorf("Check(rules = [%s], no options) error = %v, want nil", idempotencyRuleID, err)
	}
}

func runCheck(t *testing.T, rs *checktest.RequestSpec) error {
	t.Helper()
	ctx := context.Background()
	req, err := rs.ToRequest(ctx)
	if err != nil {
		t.Fatalf("Setup failed: building the request: %v", err)
	}
	client, err := check.NewClientForSpec(spec)
	if err != nil {
		t.Fatalf("Setup failed: creating the client: %v", err)
	}
	_, err = client.Check(ctx, req)
	return err
}
