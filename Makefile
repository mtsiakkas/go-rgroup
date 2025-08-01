.PHONY: test
test:
ifeq (, $(shell which tparse 2>/dev/null))
	@go test -cover -coverprofile=c.out
else
	@go test -json -cover -coverprofile=c.out | tparse -pass
endif

.PHONY: test.report
test.report: test
	@go tool cover -html=c.out

.PHONY: lint
lint:
ifeq (, $(shell which golangci-lint 2>/dev/null))
	@echo "golangci-lint not installed"
else
	golangci-lint run
endif

.PHONY: ci
ci: test lint

.PHONY: tag
TARGET?=patch
tag:
ifeq (, $(shell which git-semver 2>/dev/null))
	@echo "git-semver is required for tagging"
else
	$(eval TAG=$(shell git-semver -target $(TARGET) -prefix v -no-meta))
	@echo $(TAG)
endif

.PHONY: tag.apply
tag.apply: tag
	@git tag -s $(TAG) -m $(TAG)
