all: lint test
.PHONY: all

lint:
	go vet .
.PHONY: lint

test:
	go test .
.PHONY: test
