default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt: fmt-tf-conf
	 gofumpt -l -w .
	 gci write --skip-generated --skip-vendor -s standard -s default -s "prefix(gitlab.numspot.cloud/cloud/terraform-provider-numspot)" -s blank -s dot .

fmt-tf-conf:
	find . | egrep "test.go" | sort | while read f; do terrafmt fmt -fv $$f; done
	terraform fmt -recursive examples/

lint-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run --fix -v

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v