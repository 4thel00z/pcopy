GO=$(shell which go)
VERSION := $(shell git describe --tag)

.PHONY: help check test coverage coverage-html coverage-upload fmt fmt-check vet lint staticcheck build build-snapshot build-simple clean release release-snapshot install install-deb

## Display this help message
help: Makefile
	@echo
	@echo " Choose a command to run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## Run all tests, vetting/formatting checks and linters
check: test fmt-check vet lint staticcheck

## Run tests
test:
	$(GO) test ./...

## Run tests and show coverage
coverage:
	mkdir -p build/coverage
	$(GO) test -race -coverprofile=build/coverage/coverage.txt -covermode=atomic ./...
	$(GO) tool cover -func build/coverage/coverage.txt

## Run tests and show coverage (as HTML)
coverage-html:
	mkdir -p build/coverage
	$(GO) test -race -coverprofile=build/coverage/coverage.txt -covermode=atomic ./...
	$(GO) tool cover -html build/coverage/coverage.txt

## Upload coverage results to codecov.io
coverage-upload:
	cd build/coverage && (curl -s https://codecov.io/bash | bash)

## Run 'go fmt'
fmt:
	$(GO) fmt ./...

## Run 'go fmt', but don't change anything
fmt-check:
	test -z $(shell gofmt -l .)

## Run 'go vet'
vet:
	$(GO) vet ./...

## Run 'golint'
lint:
	which golint || $(GO) install golang.org/x/lint/golint@latest
	$(GO) list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

## Run 'staticcheck'
staticcheck:
	rm -rf build/staticcheck
	which staticcheck || $(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	mkdir -p build/staticcheck
	ln -s "$(GO)" build/staticcheck/go
	PATH="$(PWD)/build/staticcheck:$(PATH)" staticcheck ./...
	rm -rf build/staticcheck

## Build
build:
	goreleaser build --rm-dist

## Build snapshot
build-snapshot:
	goreleaser build --snapshot --rm-dist

## Build (using go build, without goreleaser)
build-simple: clean
	mkdir -p dist/pcopy_linux_amd64
	$(GO) build \
		-o dist/pcopy_linux_amd64/pcopy \
		-ldflags \
		"-s -w -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD) -X main.date=$(shell date +%s)"

## Clean build folder
clean:
	rm -rf dist build

## Create a release
release:
	goreleaser release --rm-dist

## Create a test release
release-snapshot:
	goreleaser release --snapshot --skip-publish --rm-dist

## Copy binary from dist/ to /usr/bin
install:
	sudo rm -f /usr/bin/pcopy /usr/bin/pcp /usr/bin/ppaste
	sudo cp -a dist/pcopy_linux_amd64/pcopy /usr/bin/pcopy
	sudo ln -s /usr/bin/pcopy /usr/bin/pcp
	sudo ln -s /usr/bin/pcopy /usr/bin/ppaste

## Install .deb from dist/
install-deb:
	sudo systemctl stop pcopy || true
	sudo apt-get purge pcopy || true
	sudo dpkg -i dist/*.deb
