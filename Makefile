.PHONY: run testdata init test

run:
	go run ./cmd/omnikanji

testdata:
	go run ./cmd/testdata

test:

init:
	git config core.hooksPath .githooks

