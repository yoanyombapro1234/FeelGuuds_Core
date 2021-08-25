.PHONY: proto
proto: contracts
	@echo "setting up service schema definition via protobuf"
	protoc -I. \
			-I$(GOPATH)/src \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm \
			-I=$(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit \
			-I=$(GOPATH)/src/github.com/lyft/protoc-gen-validate/validate/validate.proto \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm/options \
			-I=$(GOPATH)/src/github.com/protobuf/src/google/protobuf/timestamp.proto \
			--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
            --govalidators_out=./schema/models/ \
			--go_out=plugins=grpc:./schema/models/  --gorm_out="engine=postgres:./schema/models/" ./schema/proto/models/*.proto
.PHONY: contracts
contracts:
		@echo "setting up service level contracts with the other microservices via proto definitions"
		protoc -I. \
    			-I=$(GOPATH)/src \
    			--go_out=plugins=grpc:./schema/models/ ./schema/proto/contracts/*.proto

run-tests:
	@echo "running core-library tests"
	@echo "----- running core-metrics tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-metrics
	@echo "running core-library tests"
	@echo "----- running core-logging tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-logging
	@echo "----- running core-auth-sdk tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-auth-sdk
	@echo "----- running core-grpc tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-grpc
	@echo "----- running core-middleware tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-middleware
	@echo "----- running core-pool tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-pool
	@echo "----- running core-tlsCert tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-tlsCert
	@echo "----- running core-tracing tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-tracing/datadog
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-tracing/jaeger
	@echo "----- running core-utilities tests"
	go test github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-utilities
