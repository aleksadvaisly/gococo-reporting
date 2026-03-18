APP := gococo-reporting
BIN := bin/$(APP)
PKGS := ./...

.DEFAULT_GOAL := help

.PHONY: help build test fmt tidy clean run

help: ## show available targets
	@awk 'BEGIN {FS = ": ## "}; /^[a-zA-Z0-9_-]+: ## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build the CLI into bin/
	@mkdir -p bin
	go build -o $(BIN) ./cmd/$(APP)

test: ## run all tests
	go test $(PKGS)

fmt: ## format Go code
	gofmt -w $(shell find . -name '*.go' -type f)

tidy: ## tidy module files
	go mod tidy

clean: ## remove build artifacts
	rm -rf bin coverage.out

run: ## run the CLI; set INPUT=path[,path2] OUTPUT=report.html TITLE="My Report"
	go run ./cmd/$(APP) -input "$(INPUT)" -output "$(or $(OUTPUT),report.html)" -title "$(or $(TITLE),GoCoCo Report)"

