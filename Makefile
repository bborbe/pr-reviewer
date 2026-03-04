include Makefile.variables
include Makefile.precommit

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/bborbe/pr-reviewer/pkg/version.Version=$(VERSION)

.PHONY: run
run:
	go run -mod=mod -ldflags "$(LDFLAGS)" main.go $(ARGS)
