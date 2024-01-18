default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

lint-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55.2 golangci-lint run --fix -v

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v