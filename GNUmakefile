default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt: fmt-tf-conf
	 gofumpt -l -w .
	 gci write --skip-generated -s standard -s default -s "prefix(gitlab.numspot.cloud/cloud/terraform-provider-numspot)" -s blank -s dot .

fmt-tf-conf:
	find . | egrep "test.go" | sort | while read f; do terrafmt fmt -fv $$f; done

lint-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55.2 golangci-lint run --fix -v

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v