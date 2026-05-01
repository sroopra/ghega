package channel

import (
	"testing"
)

func TestRunFHIRTest_Pass(t *testing.T) {
	fixture := TestFixture{
		Name: "adt_a01_to_fhir",
		Input: `MSH|^~\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101000000||ADT^A01|12345|P|2.5EVN|A01|20240101000000|||PID|1||SYNTH-MRN-001^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||`,
		ExpectedJSON: `{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient","identifier":[{"value":"SYNTH-MRN-001"}],"name":[{"family":"TESTPATIENT","given":["SYNTHETIC"]}],"birthDate":"1980-01-01","gender":"male"}}]}`,
	}

	result, err := RunFHIRTest(fixture)
	if err != nil {
		t.Fatalf("RunFHIRTest error: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected test to pass, got errors: %v", result.Errors)
	}
}

func TestRunFHIRTest_MissingKey(t *testing.T) {
	fixture := TestFixture{
		Name: "adt_a01_missing_gender",
		Input: `MSH|^~\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101000000||ADT^A01|12345|P|2.5EVN|A01|20240101000000|||PID|1||SYNTH-MRN-001^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||`,
		ExpectedJSON: `{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient","gender":"female"}}]}`,
	}

	result, err := RunFHIRTest(fixture)
	if err != nil {
		t.Fatalf("RunFHIRTest error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail due to gender mismatch")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "expected female, got male") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected gender mismatch error, got: %v", result.Errors)
	}
}

func TestRunFHIRTest_ExtraKey(t *testing.T) {
	fixture := TestFixture{
		Name: "adt_a01_extra_unexpected",
		Input: `MSH|^~\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101000000||ADT^A01|12345|P|2.5EVN|A01|20240101000000|||PID|1||SYNTH-MRN-001^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||`,
		ExpectedJSON: `{"resourceType":"Bundle","type":"collection"}`,
	}

	result, err := RunFHIRTest(fixture)
	if err != nil {
		t.Fatalf("RunFHIRTest error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail due to unexpected keys")
	}
}

func TestRunFHIRTest_MissingExpectedJSON(t *testing.T) {
	fixture := TestFixture{
		Name:     "no_expected_json",
		Input:    `MSH|^~\&|S|F|R|F|20240101000000||ADT^A01|1|P|2.5\r`,
		Expected: map[string]string{"foo": "bar"},
	}

	result, err := RunFHIRTest(fixture)
	if err != nil {
		t.Fatalf("RunFHIRTest error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail because ExpectedJSON is missing")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "FHIR tests require ExpectedJSON") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'FHIR tests require ExpectedJSON' error, got: %v", result.Errors)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
