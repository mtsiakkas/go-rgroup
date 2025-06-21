.PHONY: test
test:
ifeq (, $(shell which tparse 2>/dev/null))
	@go test ./... -tags=test -short -cover
else
	@go test -json -cover ./... | tparse -pass
endif

TARGET?=patch
BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)
tag:
ifeq (, $(shell which git-semver 2>/dev/null))
	@echo "git-semver is required for tagging"
else

	$(eval TAG=$(shell git-semver -target $(TARGET) -prefix v -no-meta))
	@echo $(TAG)
endif

tag.apply: tag
	@git tag $(TAG)
