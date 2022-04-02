ALLDOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) \
                              -type f | sort)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )

.PHONY: build
build:
	goreleaser build --single-target --snapshot --rm-dist

# tool-related commands
.PHONY: install-tools
install-tools:
	go install github.com/mgechev/revive@v1.2.0
	go install github.com/google/addlicense@v1.0.0
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/client9/misspell/cmd/misspell@v0.3.4
	go install github.com/sigstore/cosign/cmd/cosign@v1.5.2
	go install github.com/goreleaser/goreleaser@v1.6.3
	go install github.com/securego/gosec/v2/cmd/gosec@v2.10.0
	go install github.com/uw-labs/lichen@v0.1.5
	go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen@v0.47.0
	

.PHONY: lint
lint:
	revive -config .revive.toml -formatter friendly ./...

.PHONY: misspell
misspell:
	misspell $(ALLDOC)

.PHONY: misspell-fix
misspell-fix:
	misspell -w $(ALLDOC)

.PHONY: test
test:
	$(MAKE) for-all CMD="go test -race ./..."

.PHONY: test-with-cover
test-with-cover:
	$(MAKE) for-all CMD="go test -race -coverprofile=cover.out ./..."
	$(MAKE) for-all CMD="go tool cover -html=cover.out -o cover.html"

.PHONY: bench
bench:
	$(MAKE) for-all CMD="go test -benchmem -run=^$$ -bench ^* ./..."

.PHONY: check-fmt
check-fmt:
	goimports -d ./ | diff -u /dev/null -

.PHONY: fmt
fmt:
	goimports -w .

.PHONY: tidy
tidy:
	$(MAKE) for-all CMD="go mod tidy -compat=1.18"

.PHONY: gosec
gosec:
	gosec ./...

# This target performs all checks that CI will do (excluding the build itself)
.PHONY: ci-checks
ci-checks: check-fmt misspell lint gosec test

.PHONY: release
release:
	goreleaser release --parallelism 4 --rm-dist

.PHONY: release-test
release-test:
	goreleaser release --parallelism 4 --skip-validate --skip-publish --skip-sign --rm-dist --snapshot

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done
