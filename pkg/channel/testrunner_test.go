package channel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sroopra/ghega/pkg/mapping"
)

func TestLoadTestFixtures_InlineInput(t *testing.T) {
	chDir := t.TempDir()
	chPath := filepath.Join(chDir, "channel.yaml")
	if err := os.WriteFile(chPath, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	tests := []Test{
		{
			Name:     "inline",
			Input:    "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
			Expected: map[string]string{"patient_mrn": "MRN12345"},
		},
	}

	fixtures, err := LoadTestFixtures(chPath, tests)
	if err != nil {
		t.Fatalf("LoadTestFixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	if fixtures[0].Input != tests[0].Input {
		t.Errorf("Input = %q, want %q", fixtures[0].Input, tests[0].Input)
	}
}

func TestLoadTestFixtures_FileInput(t *testing.T) {
	chDir := t.TempDir()
	chPath := filepath.Join(chDir, "channel.yaml")
	hl7Path := filepath.Join(chDir, "testdata", "sample.hl7")
	if err := os.MkdirAll(filepath.Dir(hl7Path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	hl7Data := "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r"
	if err := os.WriteFile(hl7Path, []byte(hl7Data), 0644); err != nil {
		t.Fatalf("write hl7: %v", err)
	}
	if err := os.WriteFile(chPath, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	tests := []Test{
		{
			Name:     "from-file",
			Input:    "testdata/sample.hl7",
			Expected: map[string]string{"patient_mrn": "MRN12345"},
		},
	}

	fixtures, err := LoadTestFixtures(chPath, tests)
	if err != nil {
		t.Fatalf("LoadTestFixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	if fixtures[0].Input != hl7Data {
		t.Errorf("Input = %q, want %q", fixtures[0].Input, hl7Data)
	}
}

func TestLoadTestFixtures_JSONFileInput(t *testing.T) {
	chDir := t.TempDir()
	chPath := filepath.Join(chDir, "channel.yaml")
	jsonPath := filepath.Join(chDir, "testdata", "input.json")
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	jsonData := `{"resourceType":"Patient","name":[{"family":"TEST"}]}`
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(chPath, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	tests := []Test{
		{
			Name:     "json-from-file",
			Input:    "testdata/input.json",
			Expected: map[string]string{},
		},
	}

	fixtures, err := LoadTestFixtures(chPath, tests)
	if err != nil {
		t.Fatalf("LoadTestFixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	if fixtures[0].Input != jsonData {
		t.Errorf("Input = %q, want %q", fixtures[0].Input, jsonData)
	}
}

func TestLoadTestFixtures_ExpectedJSONFile(t *testing.T) {
	chDir := t.TempDir()
	chPath := filepath.Join(chDir, "channel.yaml")
	jsonPath := filepath.Join(chDir, "testdata", "expected.json")
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	jsonData := `{"patient_mrn":"MRN12345"}`
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(chPath, []byte("name: test\n"), 0644); err != nil {
		t.Fatalf("write channel: %v", err)
	}

	tests := []Test{
		{
			Name:         "expected-json-from-file",
			Input:        "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
			ExpectedJSON: "testdata/expected.json",
		},
	}

	fixtures, err := LoadTestFixtures(chPath, tests)
	if err != nil {
		t.Fatalf("LoadTestFixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	if fixtures[0].ExpectedJSON != jsonData {
		t.Errorf("ExpectedJSON = %q, want %q", fixtures[0].ExpectedJSON, jsonData)
	}
	if fixtures[0].ExpectedObject == nil {
		t.Fatal("ExpectedObject should not be nil")
	}
	if fixtures[0].ExpectedObject["patient_mrn"] != "MRN12345" {
		t.Errorf("ExpectedObject[patient_mrn] = %v, want MRN12345", fixtures[0].ExpectedObject["patient_mrn"])
	}
}

func TestRunTest_Pass(t *testing.T) {
	fixture := TestFixture{
		Name:  "basic",
		Input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		Expected: map[string]string{
			"patient_mrn": "MRN12345",
		},
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected test to pass, got errors: %v", result.Errors)
	}
	if result.Actual["patient_mrn"] != "MRN12345" {
		t.Errorf("Actual patient_mrn = %q, want %q", result.Actual["patient_mrn"], "MRN12345")
	}
}

func TestRunTest_WrongValue(t *testing.T) {
	fixture := TestFixture{
		Name:  "wrong-value",
		Input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		Expected: map[string]string{
			"patient_mrn": "WRONG",
		},
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail")
	}
	wantErr := `expected "patient_mrn" = "WRONG", got "MRN12345"`
	found := false
	for _, e := range result.Errors {
		if e == wantErr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error %q, got %v", wantErr, result.Errors)
	}
}

func TestRunTest_MissingKey(t *testing.T) {
	fixture := TestFixture{
		Name:  "missing-key",
		Input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		Expected: map[string]string{
			"missing": "x",
		},
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail")
	}
	wantErr := `expected "missing" = "x", got missing key`
	found := false
	for _, e := range result.Errors {
		if e == wantErr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error %q, got %v", wantErr, result.Errors)
	}
}

func TestRunTest_ExtraKeyWarning(t *testing.T) {
	fixture := TestFixture{
		Name:     "extra-key",
		Input:    "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		Expected: map[string]string{},
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected test to pass, got errors: %v", result.Errors)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
	wantWarn := `extra key "patient_mrn" = "MRN12345"`
	if result.Warnings[0] != wantWarn {
		t.Errorf("warning = %q, want %q", result.Warnings[0], wantWarn)
	}
}

func TestRunTest_MappingEngineError(t *testing.T) {
	fixture := TestFixture{
		Name:     "engine-error",
		Input:    "not-hl7",
		Expected: map[string]string{"x": "y"},
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "x", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestJSONDiff_Equal(t *testing.T) {
	actual := map[string]any{
		"patient": map[string]any{
			"name": []any{
				map[string]any{"family": "TESTPATIENT"},
			},
		},
	}
	expected := map[string]any{
		"patient": map[string]any{
			"name": []any{
				map[string]any{"family": "TESTPATIENT"},
			},
		},
	}
	diffs := jsonDiff("", actual, expected)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got: %v", diffs)
	}
}

func TestJSONDiff_MissingKey(t *testing.T) {
	actual := map[string]any{
		"patient": map[string]any{},
	}
	expected := map[string]any{
		"patient": map[string]any{
			"name": "TESTPATIENT",
		},
	}
	diffs := jsonDiff("", actual, expected)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(diffs), diffs)
	}
	want := "patient.name: expected TESTPATIENT, got missing key"
	if diffs[0] != want {
		t.Errorf("diff = %q, want %q", diffs[0], want)
	}
}

func TestJSONDiff_NestedDiff(t *testing.T) {
	actual := map[string]any{
		"patient": map[string]any{
			"name": []any{
				map[string]any{"family": "OTHER"},
			},
		},
	}
	expected := map[string]any{
		"patient": map[string]any{
			"name": []any{
				map[string]any{"family": "TESTPATIENT"},
			},
		},
	}
	diffs := jsonDiff("", actual, expected)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(diffs), diffs)
	}
	want := `patient.name[0].family: expected TESTPATIENT, got OTHER`
	if diffs[0] != want {
		t.Errorf("diff = %q, want %q", diffs[0], want)
	}
}

func TestJSONDiff_ExtraKey(t *testing.T) {
	actual := map[string]any{
		"patient": map[string]any{
			"name":  "TESTPATIENT",
			"extra": "value",
		},
	}
	expected := map[string]any{
		"patient": map[string]any{
			"name": "TESTPATIENT",
		},
	}
	diffs := jsonDiff("", actual, expected)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %v", len(diffs), diffs)
	}
	want := "patient.extra: unexpected key with value value"
	if diffs[0] != want {
		t.Errorf("diff = %q, want %q", diffs[0], want)
	}
}

func TestRunTest_JSONDeepComparison(t *testing.T) {
	fixture := TestFixture{
		Name:  "json-equal",
		Input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		ExpectedObject: map[string]any{
			"patient_mrn": "MRN12345",
		},
		ExpectedJSON: `{"patient_mrn":"MRN12345"}`,
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected test to pass, got errors: %v", result.Errors)
	}
}

func TestRunTest_JSONMissingKey(t *testing.T) {
	fixture := TestFixture{
		Name:  "json-missing-key",
		Input: "MSH|^~\\&|GhegaApp|GhegaFac\rPID|1||MRN12345\r",
		ExpectedObject: map[string]any{
			"patient_mrn": "MRN12345",
			"missing":     "x",
		},
		ExpectedJSON: `{"patient_mrn":"MRN12345","missing":"x"}`,
	}
	mappings := []mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
	}

	result, err := RunTest(fixture, mappings)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}
	if result.Passed {
		t.Fatal("expected test to fail")
	}
	want := "missing: expected x, got missing key"
	found := false
	for _, e := range result.Errors {
		if e == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error %q, got %v", want, result.Errors)
	}
}
