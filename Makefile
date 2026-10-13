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

# Release archives: the dist target builds one per os/arch pair into DIST_DIR
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DIST_DIR  ?= dist
PLATFORMS ?= linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

.DEFAULT_GOAL := build

.PHONY: help build dist csv run install test cover vet fmt fmt-check check tidy clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build fpv-osd into bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -o $(BIN_DIR)/ ./cmd/...

dist: ## Cross-compile release archives for every PLATFORMS pair into dist/
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		name=fpv-osd_$(VERSION:v%=%)_$${os}_$${arch}; \
		exe=fpv-osd; if [ "$$os" = windows ]; then exe=fpv-osd.exe; fi; \
		echo "building $$name"; \
		mkdir -p $(DIST_DIR)/$$name; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -ldflags '-s -w' -o $(DIST_DIR)/$$name/$$exe ./cmd/fpv-osd; \
		cp LICENSE README.md $(DIST_DIR)/$$name/; \
		if [ "$$os" = windows ]; then \
			(cd $(DIST_DIR) && zip -qr $$name.zip $$name); \
		else \
			tar --owner=0 --group=0 --numeric-owner -C $(DIST_DIR) -czf $(DIST_DIR)/$$name.tar.gz $$name; \
		fi; \
		rm -rf $(DIST_DIR)/$$name; \
	done
	@cd $(DIST_DIR) && shasum -a 256 * > SHA256SUMS

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
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out
