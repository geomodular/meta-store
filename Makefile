
.DEFAULT_GOAL = help


PROTO_OPTIONS_FILES = dataset_options.proto
PROTO_SERVICE_FILES = dataset_service.proto
PROTO_PLAIN_FILES = dataset.proto
PROTO_META_FILES = dataset.proto

# Proto files; precedence matters
PROTO_FILES = $(call to_proto, $(PROTO_OPTIONS_FILES)) $(call to_proto, $(PROTO_PLAIN_FILES)) $(call to_proto, $(PROTO_SERVICE_FILES))
PB_GO_FILES = $(call to_pb_go, $(PROTO_FILES))
PB_META_GO_FILES = $(call to_pb_meta_go, $(PROTO_META_FILES))

# Functions to nicely convert paths
to_proto = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:.pb.go=.proto)))
to_proto_from_meta = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:.pb.meta.go=.proto)))
to_pb_go = $(addprefix gen/ai/h2o/meta_store/, $(notdir $(1:.proto=.pb.go)))
to_pb_meta_go = $(addprefix gen/ai/h2o/meta_store/, $(notdir $(1:.proto=.pb.meta.go)))

GOOGLE_RPC_PROTO_FILES = api/third-party/google/rpc/status.proto
GOOGLE_API_PROTO_FILES = api/third-party/google/api/annotations.proto api/third-party/google/api/field_behavior.proto api/third-party/google/api/http.proto
GOOGLE_PROTO_FILES = $(GOOGLE_RPC_PROTO_FILES) $(GOOGLE_API_PROTO_FILES)

$(GOOGLE_RPC_PROTO_FILES): | api/third-party/google/rpc
	 wget -O $@ https://raw.githubusercontent.com/googleapis/googleapis/master/$(subst api/third-party/,,$@)

$(GOOGLE_API_PROTO_FILES): | api/third-party/google/api
	 wget -O $@ https://raw.githubusercontent.com/googleapis/googleapis/master/$(subst api/third-party/,,$@)

api/third-party/google/rpc api/third-party/google/api:
	mkdir -p $@

gen:
	mkdir -p $@


.SECONDEXPANSION:
$(PB_GO_FILES): $$(call to_proto, $$@) $(PROTO_FILES) $(GOOGLE_PROTO_FILES) | gen
	protoc \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--go_out gen/ \
		--go_opt plugins=grpc \
		--go_opt paths=source_relative \
		$<


.SECONDEXPANSION:
$(PB_META_GO_FILES): $$(call to_proto_from_meta, $$@) $(PROTO_FILES) protoc-gen-meta $(GOOGLE_PROTO_FILES) | gen
	protoc \
		--plugin protoc-gen-meta \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--meta_out gen/ \
		--meta_opt paths=source_relative \
		$<


.PHONY: generate
generate: $(PB_GO_FILES) protoc-gen-meta $(PB_META_GO_FILES)  ## Generates `.go` files from `.proto`. (mandatory step)

GO_FILES = $(shell find . -name '*.go')
server: $(GO_FILES)  ## Builds meta server.
	CGO_ENABLED=0 go build -o $@ cmd/meta-store/main.go

protoc-gen-meta: cmd/meta-gen/main.go
	CGO_ENABLED=0 go build -o $@ $<

.PHONY: test
test:  ## Run tests. (needs a running server and clean database)
	go test ./... -count=1

.PHONY: clean
clean:  ## Cleans all generated and compiled files.
	rm -rf api/third-party
	rm -rf gen
	rm -f protoc-gen-meta
	rm -f server

# Utilities

print-%: ## Debug tool. Usage: make print-<var>
	@echo $* = $($*)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
