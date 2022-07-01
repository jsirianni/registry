ALLDOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) \
                              -type f | sort)

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
	go test -cover -race ./...

.PHONY: test-with-cover
test-with-cover:
	go test -race -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: check-fmt
check-fmt:
	goimports -d ./ | diff -u /dev/null -

.PHONY: fmt
fmt:
	goimports -w .

.PHONY: tidy
tidy:
	go mod tidy -compat=1.18

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

.PHONY: start-test-integration-server
start-test-integration-server:
	bash example/scripts/generate-dev-certificates.sh
	$(MAKE) build
	sudo dist/registry_linux_amd64/registry \
		--providers-dir server/testdata/providers \
		--certificate example/scripts/tls/test.crt \
		--private-key example/scripts/tls/test.key \
		--port 443

.PHONY: run-test-integration
run-test-integration:
	cd example && docker build . -t registry-client:latest
	docker run -it \
		--network host \
		registry-client:latest init
