.PHONY: clean lint test generate build install serve-issues serve-discussions verify verify-issues verify-discussions

export CGO_ENABLED=0

default: clean lint test generate build verify

clean:
	rm -rf cover.out

generate:
	go generate ./...

build: clean
	go build -ldflags "-s -w" -trimpath

install: clean generate
	go install -ldflags "-s -w" -trimpath

test: clean
	go test -v -cover ./...

lint:
	golangci-lint run

serve-issues: generate
	go run . --dir testdata/ISSUE_TEMPLATE/

serve-discussions: generate
	go run . --dir testdata/DISCUSSION_TEMPLATE

verify: verify-issues verify-discussions

verify-issues: generate
	go run . verify --dir testdata/ISSUE_TEMPLATE/

verify-discussions: generate
	go run . verify --dir testdata/DISCUSSION_TEMPLATE/
