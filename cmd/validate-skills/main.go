// validate-skills is a CLI tool that validates all skills under ai/skills/.
package main

import (
	"fmt"
	"os"

	"github.com/sroopra/ghega/internal/skills/validate"
)

func main() {
	skillsDir := "ai/skills"
	if len(os.Args) > 1 {
		skillsDir = os.Args[1]
	}

	v := validate.New(skillsDir)
	results, err := v.ValidateAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "validation error: %v\n", err)
		os.Exit(1)
	}

	var failed int
	for _, r := range results {
		if !r.Valid() {
			failed++
			fmt.Printf("FAIL %s\n", r.Path)
			for _, e := range r.Errors {
				fmt.Printf("  - %s\n", e)
			}
		} else {
			fmt.Printf("PASS %s\n", r.Path)
		}
		for _, w := range r.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	if failed > 0 {
		fmt.Printf("\n%d skill(s) failed validation\n", failed)
		os.Exit(1)
	}
	fmt.Println("\nAll skills passed validation")
}
