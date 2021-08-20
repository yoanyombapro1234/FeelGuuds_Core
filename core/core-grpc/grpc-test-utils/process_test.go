package grpc_test_utils

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yoanyombapro1234/FeelGuuds/src/services/merchant_service/gen/github.com/yoanyombapro1234/FeelGuuds/src/merchant_service/proto/merchant_service_proto_v1"
	"github.com/yoanyombapro1234/FeelGuuds/src/services/merchant_service/pkg/constants"
	grpc_utils "github.com/yoanyombapro1234/FeelGuuds/src/services/merchant_service/pkg/grpc-utils"
	"google.golang.org/grpc"
)

var server GrpcInProcessingServer

func serverStart(serviceName, collectorEndpoint string) {
	builder := GrpcInProcessingServerBuilder{}
	logger := InitializeLoggingEngine(context.Background())
	tracingEngine, closer := InitializeTracingEngine(serviceName, collectorEndpoint)
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			logger.Error(err, err.Error())
		}
	}(closer)

	builder.SetUnaryInterceptors(grpc_utils.GetDefaultUnaryServerInterceptors(tracingEngine.Tracer))

	server = builder.Build()
	server.RegisterService(func(server *grpc.Server) {
		merchant_service_proto_v1.RegisterMerchantServiceServer(server, &MockedService{})
	})
	server.Start()
}

// TestSayHello will test the HelloWorld service using A in memory data transfer instead of the normal networking
func TestSayHello(t *testing.T) {
	serverStart()
	ctx := context.Background()
	clientConn, err := GetInProcessingClientConn(ctx, server.GetListener(), []grpc.DialOption{})
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer clientConn.Close()

	client := merchant_service_proto_v1.NewMerchantServiceClient(clientConn)
	request := &merchant_service_proto_v1.CreateAccountRequest{Account: &merchant_service_proto_v1.MerchantAccount{}}
	resp, err := client.CreateAccount(ctx, request)
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}
	server.Cleanup()
	clientConn.Close()
	assert.Equal(t, resp.AccountId, uint64(1000))
}
