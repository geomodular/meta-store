
.DEFAULT_GOAL = help


# Edit this {
PROTO_CUSTOM_OPTIONS_FILES = custom_options.proto
PROTO_FILES = dataset.proto
# }

CUSTOM_OPTIONS_PB_GO_FILES = $(call to_pb_go, $(PROTO_CUSTOM_OPTIONS_FILES))
PROTO_SERVICE_FILES = $(call to_service_proto, $(PROTO_FILES))
PB_GRPC_GO_FILES = $(call to_pb_go, $(PROTO_FILES)) $(call to_pb_go, $(PROTO_SERVICE_FILES))
PB_META_GO_FILES = $(call to_pb_meta_go, $(PROTO_FILES)) $(call to_pb_meta_go, $(PROTO_SERVICE_FILES))

# Functions to nicely convert paths
to_proto = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:.pb.go=.proto)))
to_proto_from_meta = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:.pb.meta.go=.proto)))
to_proto_from_service = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:_service.proto=.proto)))
to_pb_go = $(addprefix gen/ai/h2o/meta_store/, $(notdir $(1:.proto=.pb.go)))
to_pb_meta_go = $(addprefix gen/ai/h2o/meta_store/, $(notdir $(1:.proto=.pb.meta.go)))
to_service_proto = $(addprefix api/proto/ai/h2o/meta_store/, $(notdir $(1:.proto=_service.proto)))

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
$(CUSTOM_OPTIONS_PB_GO_FILES): $$(call to_proto, $$@) $(GOOGLE_PROTO_FILES) | gen
	protoc \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--go_out gen/ \
		--go_opt plugins=grpc \
		--go_opt paths=source_relative \
		$<


.SECONDEXPANSION:
$(PB_GRPC_GO_FILES): $$(call to_proto, $$@) $(GOOGLE_PROTO_FILES) | gen
	protoc \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--go_out gen/ \
		--go_opt plugins=grpc \
		--go_opt paths=source_relative \
		$<


.SECONDEXPANSION:
$(PB_META_GO_FILES): $$(call to_proto_from_meta, $$@) protoc-gen-meta $(GOOGLE_PROTO_FILES) | gen
	protoc \
		--plugin protoc-gen-meta \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--meta_out gen/ \
		--meta_opt paths=source_relative \
		$<


.SECONDEXPANSION:
$(PROTO_SERVICE_FILES): $$(call to_proto_from_service, $$@) protoc-gen-service $(GOOGLE_PROTO_FILES) | gen
	protoc \
		--plugin protoc-gen-service \
		-I /usr/local/include \
		-I api/third-party \
		-I api/proto \
		--service_out api/proto \
		--service_opt paths=source_relative \
		$<


.PHONY: generate
generate: $(CUSTOM_OPTIONS_PB_GO_FILES) protoc-gen-service $(PROTO_SERVICE_FILES) $(PB_GRPC_GO_FILES) protoc-gen-meta $(PB_META_GO_FILES)  ## Generates `.go` files from `.proto`. (mandatory step)

GO_FILES = $(shell find . -name '*.go')
server: $(GO_FILES)  ## Builds meta server.
	CGO_ENABLED=0 go build -o $@ cmd/meta-store/main.go

protoc-gen-meta: cmd/meta-gen/main.go
	CGO_ENABLED=0 go build -o $@ $<

protoc-gen-service: cmd/service-gen/main.go
	CGO_ENABLED=0 go build -o $@ $<

.PHONY: test
test:  ## Run tests. (needs a running server and clean database)
	go test ./... -count=1

.PHONY: clean
clean:  ## Cleans all generated and compiled files.
	rm -rf gen
	rm -f protoc-gen-meta
	rm -f protoc-gen-service
	rm -f server
	rm -f api/proto/ai/h2o/meta_store/*_service.proto

.PHONY: purge
purge: clean  ## Cleans all generated and compiled files along with downloaded proto files.
	rm -rf api/third-party

# Utilities

print-%: ## Debug tool. Usage: make print-<var>
	@echo $* = $($*)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
