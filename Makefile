.PHONY: run fixture init test

run:
	go run ./cmd/omnikanji

fixture:
	go run ./cmd/fixture

test:
	go test ./...

init:
	git config core.hooksPath .githooks

