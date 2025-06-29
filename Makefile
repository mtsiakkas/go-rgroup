.PHONY: test
test:
ifeq (, $(shell which tparse 2>/dev/null))
	@go test -cover -coverprofile=c.out ./... 
else
	@go test -json -cover -coverprofile=c.out ./... | tparse -pass
endif

.PHONY: test.report
test.report: test
	@go tool cover -html=c.out

TARGET?=patch
tag:
ifeq (, $(shell which git-semver 2>/dev/null))
	@echo "git-semver is required for tagging"
else
	$(eval TAG=$(shell git-semver -target $(TARGET) -prefix v -no-meta))
	@echo $(TAG)
endif

tag.apply: tag
	@git tag $(TAG)
