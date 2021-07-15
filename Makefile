# Makefile

.PHONY: all
all: build

.PHONY: lint
lint:
	@outputfile="$$(mktemp)" ; \
	go list ./... | xargs golint 2>&1 | \
		sed -e "s#^$$GOPATH/src/##" | tee "$$outputfile" ; \
	lines="$$(wc -l <"$$outputfile")" ; \
	rm -f "$$outputfile" ; \
	exit "$$lines"

.PHONY: retest
retest: ## Re-run all unit tests
	@./script/build.sh -a retest

.PHONY: test
test: ## Run tests
	@go test ./...

.PHONY: spec_feature_test
spec_feature_test: ## run integration tests in the specs internal repo
	@specsrepo="$(PWD)/../specs-internal" \
	./script/build.sh -a spec_feature_test

.PHONY: race
race: ## Run data race detector
	@./script/build.sh -a race

.PHONY: mocks
mocks: ## Make mocks
	@./script/build.sh -a mocks

.PHONY: msan
msan: ## Run memory sanitizer
	@if ! which clang 1>/dev/null ; then echo "Need clang" ; exit 1 ; fi
	@env CC=clang CGO_ENABLED=1 go test -msan ./...

.PHONY: vet
vet: deps
	@go vet ./...

.PHONY: coverage
coverage:
	@go test -covermode=count -coverprofile="coverage.txt" ./...
	@go tool cover -func="coverage.txt"

.PHONY: deps
deps: ## Get the dependencies
	@./script/build.sh -a deps

.PHONY: build
build:
	@go build .

.PHONY: gofmtsimplify
gofmtsimplify:
	@find . -path vendor -prune -o \( -name '*.go' -and -not -name '*_test.go' -and -not -name '*_mock.go' \) -print0 | xargs -0r gofmt -s -w

.PHONY: install
install: ## install the binaries in GOPATH/bin
	@./script/build.sh -a install -t default

.PHONY: gqlgen
gqlgen: ## run gqlgen
	@./script/build.sh -a gqlgen

.PHONY: gqlgen_check
gqlgen_check: ## GraphQL: Check committed files match just-generated files
	@find gateway/graphql -name '*.graphql' -o -name '*.yml' -exec touch '{}' ';' ; \
	make gqlgen 1>/dev/null || exit 1 ; \
	files="$$(git diff --name-only gateway/graphql/)" ; \
	if test -n "$$files" ; then \
		echo "Committed files do not match just-generated files:" $$files ; \
		test -n "$(CI)" && git diff gateway/graphql/ ; \
		exit 1 ; \
	fi

.PHONY: ineffectassign
ineffectassign: ## Check for ineffectual assignments
	@ia="$$(env GO111MODULE=auto ineffassign . | grep -v '_test\.go:')" ; \
	if test "$$(echo -n "$$ia" | wc -l | awk '{print $$1}')" -gt 0 ; then echo "$$ia" ; exit 1 ; fi

.PHONY: proto
proto: ## build proto definitions
	@./proto/generate.sh

.PHONY: proto_check
proto_check: ## proto: Check committed files match just-generated files
	@make proto_clean 1>/dev/null
	@make proto 1>/dev/null
	@files="$$(git diff --name-only proto/)" ; \
	if test -n "$$files" ; then \
		echo "Committed files do not match just-generated files:" $$files ; \
		test -n "$(CI)" && git diff proto/ ; \
		exit 1 ; \
	fi

.PHONY: proto_clean
proto_clean:
	@find proto -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*.validator.pb.go' -o -name '*.swagger.json' \
		| xargs -r rm

# Misc Targets

codeowners_check:
	@if grep -v '^#' CODEOWNERS | grep "," ; then \
		echo "CODEOWNERS cannot have entries with commas" ; \
		exit 1 ; \
	fi

.PHONY: print_check
print_check: ## Check for fmt.Print functions in Go code
	@f="$$(mktemp)" && \
	find -name vendor -prune -o \
		-name cmd -prune -o \
		-name '*_test.go' -prune -o \
		-name 'flags.go' -prune -o \
		-name '*.go' -print0 | \
		xargs -0 grep -E '^([^/]|/[^/])*fmt.Print' | \
		tee "$$f" && \
	count="$$(wc -l <"$$f")" && \
	rm -f "$$f" && \
	if test "$$count" -gt 0 ; then exit 1 ; fi

.PHONY: gettools_develop
gettools_develop:
	@./script/gettools.sh develop

# Make sure the mdspell command matches the one in .drone.yml.
.PHONY: spellcheck
spellcheck: ## Run markdown spellcheck container
	@docker run --rm -ti \
		--entrypoint mdspell \
		-v "$(PWD):/src" \
		docker.pkg.github.com/vegaprotocol/devops-infra/markdownspellcheck:latest \
			--en-gb \
			--ignore-acronyms \
			--ignore-numbers \
			--no-suggestions \
			--report \
			'*.md' \
			'docs/**/*.md'

# The integration directory is special, and contains a package called core_test.
.PHONY: staticcheck
staticcheck: ## Run statick analysis checks
	@./script/build.sh -a staticcheck

.PHONY: buflint
buflint: ## Run buf lint
	@./script/build.sh -a buflint

.PHONY: misspell
misspell: # Run go specific misspell checks
	@./script/build.sh -a misspell

.PHONY: semgrep
semgrep: ## Run semgrep static analysis checks
	@./script/build.sh -a semgrep

.PHONY: clean
clean: SHELL:=/bin/bash
clean: ## Remove previous build
	@source ./script/build.sh && \
	rm -f cmd/*/*.log && \
	for app in "$${apps[@]}" ; do \
		rm -f "$$app" "cmd/$$app/$$app" "cmd/$$app/$$app"-* ; \
	done

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
