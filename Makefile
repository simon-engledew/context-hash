MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

rwildcard = $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

bin/docker-context-hash: $(call rwildcard,src/,*.go)
	(cd src; go build -trimpath -ldflags="-w -s" -o $(abspath $@) .)

.PHONY: build-example
build-example: bin/docker-context-hash
	docker build ./example -t "docker-context-hash:$$(bin/docker-context-hash ./example)"

.PHONY: example
example: build-example
	docker run --rm -it "docker-context-hash:92fb07f5612143976c8a6420f2f022ce094bb3573d826f063bc4ff18f508c829"