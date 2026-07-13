package redaction

import (
	"reflect"
	"testing"
)

func TestInspectDefaultPolicyRecursesAndDoesNotLeakValues(t *testing.T) {
	findings, err := Inspect([]byte(`{"Authorization":"Bearer top-secret","fields":{"nested-value":{"PRIVATE-KEY":"-----BEGIN PRIVATE KEY----- secret"}}}`), DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	want := []Finding{
		{RuleID: "bearer_credential_value", NormalizedFieldPath: "envelope.authorization", ReasonCode: "forbidden_bearer_credential"},
		{RuleID: "forbidden_auth_transport_fields", NormalizedFieldPath: "envelope.authorization", ReasonCode: "forbidden_auth_data"},
		{RuleID: "forbidden_credential_fields", NormalizedFieldPath: "envelope.fields.nestedvalue.privatekey", ReasonCode: "forbidden_credential_data"},
		{RuleID: "private_key_pem_header", NormalizedFieldPath: "envelope.fields.nestedvalue.privatekey", ReasonCode: "forbidden_private_key_material"},
	}
	if !reflect.DeepEqual(findings, want) {
		t.Fatalf("findings = %#v, want %#v", findings, want)
	}
	for _, finding := range findings {
		if finding.RuleID == "top-secret" || finding.NormalizedFieldPath == "secret" || finding.ReasonCode == "secret" {
			t.Fatal("finding leaked rejected value")
		}
	}
}

func TestInspectFindingOrderIsDeterministic(t *testing.T) {
	input := []byte(`{"z_secret":"x","a_secret":"y","fields":[{"Cookie":"x"},{"api-key":"y"}]}`)
	first, err := Inspect(input, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	second, err := Inspect(input, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("finding order changed: %#v != %#v", first, second)
	}
	if len(first) != 2 || first[0].NormalizedFieldPath != "envelope.fields[0].cookie" || first[1].NormalizedFieldPath != "envelope.fields[1].apikey" {
		t.Fatalf("findings = %#v", first)
	}
}

func TestInspectSupportsCustomPolicy(t *testing.T) {
	policy := Policy{Separators: []rune{'_', '-'}, CaseInsensitive: true, RemoveSeparators: true, Rules: []Rule{{ID: "custom", Keys: []string{"sensitive_key"}, ReasonCode: "custom_reason"}}}
	findings, err := Inspect([]byte(`{"Sensitive-Key":123}`), policy)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].RuleID != "custom" || findings[0].NormalizedFieldPath != "envelope.sensitivekey" {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestDefaultPolicyBearerPatternMatchesWhitespace(t *testing.T) {
	findings, err := Inspect([]byte(`{"message":"  Bearer\tsecret-token  "}`), DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].RuleID != "bearer_credential_value" {
		t.Fatalf("findings = %#v", findings)
	}
}
