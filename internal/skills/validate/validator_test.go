package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidSkillPasses(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "test-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: test-skill
description: >
  Use when testing the validator.
license: Apache-2.0
---

# Test Skill

This is a test skill.
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if !result.Valid() {
		t.Fatalf("expected valid skill, got errors: %v", result.Errors)
	}
}

func TestMissingFrontmatterFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "bad-skill")
	os.MkdirAll(skillDir, 0755)

	content := `# Bad Skill

No frontmatter here.
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to missing frontmatter")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "frontmatter") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected frontmatter error, got: %v", result.Errors)
	}
}

func TestInvalidNameFormatFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "Bad_Name")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: Bad_Name
description: >
  Use when testing.
license: Apache-2.0
---

# Bad Name
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to bad name format")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "does not match") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected name format error, got: %v", result.Errors)
	}
}

func TestMissingDescriptionFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-desc")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: no-desc
license: Apache-2.0
---

# No Description
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to missing description")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "description") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected description error, got: %v", result.Errors)
	}
}

func TestPHIInExamplesFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "phi-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: phi-skill
description: >
  Use when testing PHI detection.
license: Apache-2.0
---

# PHI Example

Here is a fake SSN: 123-45-6789.
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to PHI in examples")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "SSN") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected SSN/PHI error, got: %v", result.Errors)
	}
}

func TestMissingLicenseFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-license")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: no-license
description: >
  Use when testing.
---

# No License
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to missing license")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "license") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected license error, got: %v", result.Errors)
	}
}

func TestMissingReferencedFileFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "bad-ref")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: bad-ref
description: >
  Use when testing.
license: Apache-2.0
---

# Bad Ref

See [missing](references/missing.md).
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to missing reference")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "referenced file missing") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing reference error, got: %v", result.Errors)
	}
}

func TestJavaScriptFileFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "js-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: js-skill
description: >
  Use when testing.
license: Apache-2.0
---

# JS Skill
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
	os.WriteFile(filepath.Join(skillDir, "bad.js"), []byte("alert('hi')"), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to JS file")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "JavaScript") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected JavaScript error, got: %v", result.Errors)
	}
}

func TestScriptTagFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "script-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: script-skill
description: >
  Use when testing.
license: Apache-2.0
---

# Script Skill

<script>alert('x')</script>
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to script tag")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "<script>") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected script tag error, got: %v", result.Errors)
	}
}

func TestSecretsInExamplesFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "secret-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: secret-skill
description: >
  Use when testing.
license: Apache-2.0
---

# Secret Skill

api_key=supersecret123
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to secret in example")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "secret") || contains(e, "credential") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected secret error, got: %v", result.Errors)
	}
}

func TestLineCountExceededFails(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "long-skill")
	os.MkdirAll(skillDir, 0755)

	var lines []string
	lines = append(lines, "---")
	lines = append(lines, "name: long-skill")
	lines = append(lines, "description: >")
	lines = append(lines, "  Use when testing.")
	lines = append(lines, "license: Apache-2.0")
	lines = append(lines, "---")
	lines = append(lines, "")
	lines = append(lines, "# Long Skill")
	for i := 0; i < 500; i++ {
		lines = append(lines, "line")
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	v := New(dir)
	result := v.ValidateSkill(skillDir)
	if result.Valid() {
		t.Fatal("expected invalid skill due to excessive line count")
	}
	found := false
	for _, e := range result.Errors {
		if contains(e, "500 lines") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected line count error, got: %v", result.Errors)
	}
}

func TestValidateAll(t *testing.T) {
	dir := t.TempDir()

	// Valid skill
	s1 := filepath.Join(dir, "skill-one")
	os.MkdirAll(s1, 0755)
	os.WriteFile(filepath.Join(s1, "SKILL.md"), []byte("---\nname: skill-one\ndescription: >\n  Use when testing.\nlicense: Apache-2.0\n---\n\n# One\n"), 0644)

	// Invalid skill
	s2 := filepath.Join(dir, "skill two")
	os.MkdirAll(s2, 0755)
	os.WriteFile(filepath.Join(s2, "SKILL.md"), []byte("---\nname: skill two\ndescription: >\n  Use when testing.\nlicense: Apache-2.0\n---\n\n# Two\n"), 0644)

	v := New(dir)
	results, err := v.ValidateAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	var valid, invalid int
	for _, r := range results {
		if r.Valid() {
			valid++
		} else {
			invalid++
		}
	}
	if valid != 1 || invalid != 1 {
		t.Fatalf("expected 1 valid and 1 invalid, got %d valid, %d invalid", valid, invalid)
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
