// Package validate provides skill validation logic for the ai/skills/ directory.
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Result holds the outcome of validating a single skill directory.
type Result struct {
	Path    string
	Errors  []string
	Warnings []string
}

// Valid returns true if the skill has no validation errors.
func (r Result) Valid() bool {
	return len(r.Errors) == 0
}

// Validator performs skill validation.
type Validator struct {
	root string
}

// New creates a Validator rooted at the given skills directory.
func New(root string) *Validator {
	return &Validator{root: root}
}

// ValidateAll walks the skills directory and validates each skill.
func (v *Validator) ValidateAll() ([]Result, error) {
	entries, err := os.ReadDir(v.root)
	if err != nil {
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	var results []Result
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(v.root, entry.Name())
		results = append(results, v.ValidateSkill(skillPath))
	}
	return results, nil
}

// ValidateSkill validates a single skill directory.
func (v *Validator) ValidateSkill(skillPath string) Result {
	r := Result{Path: skillPath}

	skillFile := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("SKILL.md missing or unreadable: %v", err))
		return r
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 500 {
		r.Errors = append(r.Errors, fmt.Sprintf("SKILL.md exceeds 500 lines (%d)", len(lines)))
	}

	fm, body, err := parseFrontmatter(string(data))
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("frontmatter parse error: %v", err))
		return r
	}

	if fm.Name == "" {
		r.Errors = append(r.Errors, "frontmatter missing 'name'")
	} else {
		validName := regexp.MustCompile(`^[a-z0-9-]+$`)
		if !validName.MatchString(fm.Name) {
			r.Errors = append(r.Errors, fmt.Sprintf("name %q does not match ^[a-z0-9-]+$", fm.Name))
		}
	}

	if fm.Description == "" {
		r.Errors = append(r.Errors, "frontmatter missing 'description'")
	} else if !strings.Contains(fm.Description, "Use when") {
		r.Errors = append(r.Errors, "description missing trigger phrase 'Use when'")
	}

	if fm.License == "" {
		r.Errors = append(r.Errors, "frontmatter missing 'license'")
	}

	// Check for executable JavaScript in the skill directory.
	if jsErrs := checkNoJavaScript(skillPath); len(jsErrs) > 0 {
		r.Errors = append(r.Errors, jsErrs...)
	}

	// Check for secrets and PHI in the entire skill directory content.
	if secErrs := checkSecretsAndPHI(skillPath); len(secErrs) > 0 {
		r.Errors = append(r.Errors, secErrs...)
	}

	// Check that all referenced files exist.
	if refErrs := checkReferences(skillPath, body); len(refErrs) > 0 {
		r.Errors = append(r.Errors, refErrs...)
	}

	return r
}

// Frontmatter represents the YAML frontmatter of a SKILL.md file.
type Frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	License     string `yaml:"license"`
}

// parseFrontmatter extracts YAML frontmatter and the markdown body.
func parseFrontmatter(content string) (Frontmatter, string, error) {
	var fm Frontmatter
	if !strings.HasPrefix(content, "---") {
		return fm, "", fmt.Errorf("missing frontmatter delimiter")
	}
	rest := strings.TrimPrefix(content, "---")
	parts := strings.SplitN(rest, "---", 2)
	if len(parts) != 2 {
		return fm, "", fmt.Errorf("frontmatter not properly closed")
	}

	// We avoid importing a full YAML library by using a minimal
	// line-based parser for the three known scalar keys.
	for _, line := range strings.Split(parts[0], "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			fm.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "description:") {
			// description may start on this line or continue as folded literal
			fm.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
		if strings.HasPrefix(line, "license:") {
			fm.License = strings.TrimSpace(strings.TrimPrefix(line, "license:"))
		}
	}

	// If description was indicated with `>` it will likely be on subsequent lines.
	// Re-parse with a tiny state machine to capture multi-line folded/literal blocks.
	fm = parseFrontmatterScalars(parts[0])

	body := strings.TrimSpace(parts[1])
	return fm, body, nil
}

func parseFrontmatterScalars(block string) Frontmatter {
	var fm Frontmatter
	lines := strings.Split(block, "\n")
	var currentKey string
	var currentValue strings.Builder

	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect a new key at the root level (no leading spaces, contains colon)
		if !strings.HasPrefix(raw, " ") && !strings.HasPrefix(raw, "\t") && strings.Contains(raw, ":") {
			// flush previous
			switch currentKey {
			case "name":
				fm.Name = strings.TrimSpace(currentValue.String())
			case "description":
				fm.Description = strings.TrimSpace(currentValue.String())
			case "license":
				fm.License = strings.TrimSpace(currentValue.String())
			}
			currentValue.Reset()

			parts := strings.SplitN(raw, ":", 2)
			currentKey = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				val := strings.TrimSpace(parts[1])
				if val != "" && val != ">" && val != "|" {
					currentValue.WriteString(val)
					currentValue.WriteString(" ")
				}
			}
			continue
		}

		// Continuation line for multi-line scalar
		if currentKey != "" {
			currentValue.WriteString(trimmed)
			currentValue.WriteString(" ")
		}
	}

	switch currentKey {
	case "name":
		fm.Name = strings.TrimSpace(currentValue.String())
	case "description":
		fm.Description = strings.TrimSpace(currentValue.String())
	case "license":
		fm.License = strings.TrimSpace(currentValue.String())
	}

	return fm
}

func checkNoJavaScript(skillPath string) []string {
	var errs []string
	err := filepath.Walk(skillPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".js") {
			errs = append(errs, fmt.Sprintf("executable JavaScript file found: %s", path))
		}
		if strings.HasSuffix(strings.ToLower(path), ".md") || strings.HasSuffix(strings.ToLower(path), ".html") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(strings.ToLower(string(data)), "<script") {
				errs = append(errs, fmt.Sprintf("<script> tag found in: %s", path))
			}
		}
		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Sprintf("error walking skill directory: %v", err))
	}
	return errs
}

var (
	ssnPattern     = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	phonePattern   = regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`)
	emailPattern   = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	passwordPattern = regexp.MustCompile(`(?i)(password\s*=\s*\S+|api_key\s*=\s*\S+|secret\s*=\s*\S+|token\s*=\s*\S+)`)
)

func checkSecretsAndPHI(skillPath string) []string {
	var errs []string
	err := filepath.Walk(skillPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		lower := strings.ToLower(content)

		if ssnPattern.MatchString(content) {
			errs = append(errs, fmt.Sprintf("possible SSN detected in %s", path))
		}
		if phonePattern.MatchString(content) {
			errs = append(errs, fmt.Sprintf("possible phone number detected in %s", path))
		}
		if emailPattern.MatchString(content) {
			errs = append(errs, fmt.Sprintf("possible email detected in %s", path))
		}
		if passwordPattern.MatchString(content) {
			errs = append(errs, fmt.Sprintf("possible secret/credential detected in %s", path))
		}

		// Heuristic for real patient names in HL7 PID segments.
		if strings.Contains(lower, "pid|") && !strings.Contains(lower, "testpatient") && !strings.Contains(lower, "synthetic") {
			// This is a weak heuristic; flag only if we see realistic-looking names after PID.
			if realisticNamePattern.MatchString(content) {
				errs = append(errs, fmt.Sprintf("possible real patient name in HL7 segment in %s", path))
			}
		}

		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Sprintf("error walking skill directory: %v", err))
	}
	return errs
}

// Very basic pattern to catch common first+last name shapes inside HL7 PID lines.
// This is intentionally conservative to avoid false positives on common words.
var realisticNamePattern = regexp.MustCompile(`(?i)PID\|.*\b(John|Jane|Smith|Doe|Johnson|Williams|Brown|Jones|Garcia|Miller|Davis|Rodriguez|Martinez|Hernandez|Lopez|Gonzalez|Wilson|Anderson|Thomas|Taylor|Moore|Jackson|Martin|Lee|Perez|Thompson|White|Harris|Sanchez|Clark|Ramirez|Lewis|Robinson|Walker|Young|Allen|King|Wright|Scott|Torres|Nguyen|Hill|Flores|Green|Adams|Nelson|Baker|Hall|Rivera|Campbell|Mitchell|Carter|Roberts|Gomez|Phillips|Evans|Turner|Diaz|Parker|Cruz|Edwards|Collins|Reyes|Stewart|Morris|Morales|Murphy|Cook|Rogers|Gutierrez|Ortiz|Morgan|Cooper|Peterson|Bailey|Reed|Kelly|Howard|Ramos|Kim|Cox|Ward|Richardson|Watson|Brooks|Chavez|Wood|James|Bennett|Gray|Mendoza|Ruiz|Hughes|Price|Alvarez|Castillo|Sanders|Patel|Myers|Long|Ross|Foster|Jimenez|Powell|Jenkins|Perry|Russell|Sullivan|Bell|Coleman|Butler|Henderson|Barnes|Gonzales|Fisher|Vasquez|Simmons|Romero|Jordan|Patterson|Alexander|Hamilton|Graham|Reynolds|Griffin|Wallace|Moreno|West|Cole|Hayes|Bryant|Herrera|Gibson|Ellis|Tran|Medina|Wagner|Hunter|Dixon|Muir|Rai|Sharma|Gupta|Singh|Kumar|Patel|Shah|Desai|Reddy|Nair|Iyer|Mehta|Joshi|Bhat|Rao|Pillai|Menon|Nambiar|Verma|Agarwal|Kapoor|Malhotra|Banerjee|Sen|Das|Bose|Ghosh|Mitra|Chatterjee|Banerjee|Roy|Choudhury|Mazumdar|Talukdar|Barua|Hazarika|Bora|Saikia|Goswami|Borah|Phukan|Dutta|Baruah|Sarma|Khan|Ahmed|Ali|Hassan|Malik|Rahman|Hussain|Siddiqui|Siddiqi|Akhtar|Chowdhury|Haque|Islam|Uddin|Khatun|Begum|Parveen|Yasmin|Sultana|Jahan|Nahar|Khatun|Akhter|Hossain|Mahmud|Karim|Mannan|Mia|Quader|Aziz|Kamal|Hamid|Rashid|Anwar|Ashraf|Aslam|Asghar|Iqbal|Javed|Khalid|Latif|Majeed|Masood|Mumtaz|Nadeem|Nasir|Nawaz|Qureshi|Rafiq|Rana|Rehman|Salim|Sarwar|Shafiq|Shahbaz|Shaukat|Sohail|Tahir|Wahab|Yousuf|Zahid|Zaman|Zubair)\b`)

// checkReferences scans the markdown body for relative links and ensures they resolve.
func checkReferences(skillPath, body string) []string {
	var errs []string
	// Match markdown links: [text](path)
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		ref := m[2]
		// Skip external URLs and anchors
		if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "#") {
			continue
		}
		target := filepath.Join(skillPath, ref)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("referenced file missing: %s", ref))
		}
	}
	return errs
}
