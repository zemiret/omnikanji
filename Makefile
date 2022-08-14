.PHONY: run test

run:
	go run ./cmd/omnikanji

test:

init:
	git config core.hooksPath .githooks
