include .env

## deps: go deps install
deps:
	go mod tidy && go mod download

## build：build binary
build:
	@echo "  >  Building binary..."
	@echo "  >  Use build param: ${GO_LDFLAGS}"
	@go build -ldflags "${GO_LDFLAGS}" -o $(GOBIN)/hyperops ./main.go
	@echo "  >  Building finish..."

## build_for_mac：build binary
build_for_mac:
	@echo "  >  Building binary..."
	@echo "  >  Use build param: ${GO_LDFLAGS}"
	@GOOS=darwin GOARCH=amd64 go build -ldflags "${GO_LDFLAGS}" -o $(GOBIN)/hyperops ./main.go
	@echo "  >  Building finish..."


## fmt: format all go code
fmt:
	@echo "fmt code"
	find ./pkg -name "*.go" | xargs goimports -w
	find ./cmd -name "*.go" | xargs goimports -w
	find ./internal -name "*.go" | xargs goimports -w

## lint: lint go code
lint:
	@echo "lint code"
	golangci-lint run --verbose --max-same-issues 0 --sort-results --timeout 3m --skip-dirs vendor

## clean: clean
clean:
	@echo "clean bin files"
	rm -rf ./bin/*
	rm -rf ./log/*


all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: ctags
