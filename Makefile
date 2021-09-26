.PHONY: test
test:
	@go test -race -timeout 1s ./...

.PHONY: gen
gen:
	@go generate ./...
