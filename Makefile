# Ghega — Makefile
# https://github.com/ghega/ghega

BINARY := ghega
MODULE := github.com/sroopra/ghega
IMAGE := ghcr.io/ghega/ghega:local
GOFLAGS ?= -ldflags="-s -w"

.PHONY: dev
## dev: run the Ghega CLI in development mode
dev:
	go run ./cmd/ghega

.PHONY: test
## test: run all Go tests and validation checks
test:
	go test ./...

.PHONY: lint
## lint: run go vet (and golangci-lint if available)
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; skipping (install from https://golangci-lint.run/)"; \
	fi

.PHONY: build
## build: compile the ghega binary to the repo root
build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/ghega

.PHONY: docker
## docker: build the Ghega container image
docker:
	docker build -t $(IMAGE) .

.PHONY: validate-skills
## validate-skills: placeholder for skill validation (to be wired later)
validate-skills:
	@echo "skills validation not yet implemented"

.PHONY: test-generators
## test-generators: placeholder for generator tests (to be wired later)
test-generators:
	@echo "generator tests not yet implemented"

.PHONY: test-runtime-no-java-js
## test-runtime-no-java-js: verify runtime boundary (no Java / no JavaScript)
test-runtime-no-java-js:
	bash scripts/test-runtime-no-java-js.sh

.PHONY: adr
## adr: create a new ADR from docs/adr/template.md (usage: make adr NEW="my-decision-title")
adr:
	@if [ -z "$(NEW)" ]; then \
		echo "Usage: make adr NEW=\"My Decision Title\""; \
		exit 1; \
	fi
	@mkdir -p docs/adr
	@# Find the next ADR number
	@next=$$(ls docs/adr/*.md 2>/dev/null | sed 's|.*/||' | grep -oE '^[0-9]+' | sort -n | tail -1); \
	if [ -z "$$next" ]; then next=1; else next=$$(( $$(echo "$$next" | sed 's/^0*//') + 1 )); fi; \
	num=$$(printf "%03d" $$next); \
	slug=$$(echo "$(NEW)" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$$//'); \
	file="docs/adr/$${num}-$${slug}.md"; \
	cp docs/adr/template.md "$$file"; \
	echo "Created $$file"
