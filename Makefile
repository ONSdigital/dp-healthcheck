.PHONY: all
all: audit test build

.PHONY: test
test:
	go test -race -cover ./...

audit:
	go list -json -m all | nancy sleuth
.PHONY: audit

build:
	go build ./...
.PHONY: build

lint:
	exit
.PHONY: lint
