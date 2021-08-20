package grpc_test_utils

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yoanyombapro1234/FeelGuuds/src/services/merchant_service/gen/github.com/yoanyombapro1234/FeelGuuds/src/merchant_service/proto/merchant_service_proto_v1"
	grpc_utils "github.com/yoanyombapro1234/FeelGuuds_Core/core/grpc-utils"

	"google.golang.org/grpc"
)

func startServer(serviceName, collectorEndpoint string) GrpcInProcessingServer {
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
	server := builder.Build()
	server.RegisterService(func(server *grpc.Server) {
		merchant_service_proto_v1.RegisterMerchantServiceServer(server, &MockedService{})
	})
	server.Start()
	return server
}

func TestCreateAccountEndpointPassingContext(t *testing.T) {
	server := startServer()
	ctx := context.Background()
	clientBuilder := InProcessingClientBuilder{Server: server}
	clientBuilder.WithInsecure()
	clientBuilder.WithContext(ctx)

	if server == nil {
		t.Fatalf("failed to initialize server object")
	}
	dialer := GetBufDialer(server.GetListener())
	clientBuilder.WithOptions(grpc.WithContextDialer(dialer))
	clientConn, err := clientBuilder.GetConn("localhost", "50051")
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
