GO      ?= go
BIN_DIR ?= bin

# CSV format for the csv target: ; and , suit spreadsheets with a Portuguese locale
SEP     ?= ;
DECIMAL ?= ,
FILES   ?= $(wildcard *.osd)

# Input for the run target: make run FILE=flight.osd ARGS="-sep ';' -decimal ','"
# CMD picks the subcommand: csv (default) or dump, e.g. make run CMD=dump ARGS=-census
CMD  ?= csv
FILE ?= AvatarG0031.osd
ARGS ?=

.DEFAULT_GOAL := build

.PHONY: help build csv run install test cover vet fmt fmt-check check tidy clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build fpv-osd into bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -o $(BIN_DIR)/ ./cmd/...

csv: build ## Convert every .osd in FILES to CSV (SEP and DECIMAL set the format)
	@test -n "$(FILES)" || { echo "no .osd files found (set FILES=...)"; exit 1; }
	@for f in $(FILES); do $(BIN_DIR)/fpv-osd csv -sep '$(SEP)' -decimal '$(DECIMAL)' "$$f" || exit 1; done

run: ## Run fpv-osd CMD on FILE (extra options in ARGS)
	$(GO) run ./cmd/fpv-osd $(CMD) $(ARGS) $(FILE)

install: ## Install fpv-osd into GOBIN
	$(GO) install -trimpath ./cmd/...

test: ## Run the tests
	$(GO) test ./...

cover: ## Run the tests with a coverage summary
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -n 1

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format the code with gofmt
	gofmt -w .

fmt-check: ## Fail if any file is not gofmt-formatted
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

check: fmt-check vet test ## Run all checks (format, vet, tests)

tidy: ## Tidy go.mod
	$(GO) mod tidy

clean: ## Remove build and coverage output
	rm -rf $(BIN_DIR) coverage.out
