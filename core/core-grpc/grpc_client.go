package core_grpc

import (
	"context"
	"crypto/x509"
	"time"

	grpcclient "github.com/apssouza22/grpc-production-go/client"
	"github.com/apssouza22/grpc-production-go/grpcutils"
	tlscert "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-tlsCert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

type GrpcClient struct {
	Logger *zap.Logger
	Conn   *grpc.ClientConn
}

// NewGrpcClient initializes a new GRPC client connection
// usage:
//  gc := NewGrpcClient(addr string, logger *zap.Logger, tlsEnabled bool, cert *x509.CertPool)
// defer gc.Conn.Close()
//
func (grpcClientInstance *GrpcClient) NewGrpcClient(addr string, logger *zap.Logger, tlsEnabled bool, cert *x509.CertPool) *GrpcClient {
	clientBuilder := grpcclient.GrpcConnBuilder{}
	clientBuilder.WithInsecure()
	clientBuilder.WithContext(context.Background())
	if tlsEnabled {
		if cert != nil {
			clientBuilder.WithClientTransportCredentials(false, cert)
		} else {
			clientBuilder.WithClientTransportCredentials(false, tlscert.CertPool)
		}
	}

	clientBuilder.WithStreamInterceptors(grpcutils.GetDefaultStreamClientInterceptors())
	clientBuilder.WithUnaryInterceptors(grpcutils.GetDefaultUnaryClientInterceptors())
	cc, err := clientBuilder.GetConn(addr)
	if err != nil {
		logger.Fatal(err.Error())
	}

	return &GrpcClient{
		Logger: logger,
		Conn:   cc,
	}
}

// SendRpcRequest performs an rpc request against a downstream server
func (grpcClientInstance *GrpcClient) SendRpcRequest(ctx context.Context, ctxPairs []string, rpcOp func(ctx context.Context,
	param ...interface{}) (response interface{}, err error),
	requestParams ...interface{}) (rpcResponse interface{}, err error) {
	md := metadata.Pairs(ctxPairs...)
	ctx = metadata.NewOutgoingContext(ctx, md)
	timeout := time.Minute * 1
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	{
		healthClient := grpc_health_v1.NewHealthClient(grpcClientInstance.Conn)
		response, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			grpcClientInstance.Logger.Error(err.Error())
			return nil, err
		}
		grpcClientInstance.Logger.Info("successfully obtained response from health client", zap.Any("Response", response))
	}

	rpcResponse, err = rpcOp(ctx, requestParams...)
	if err != nil {
		grpcClientInstance.Logger.Error(err.Error())
		return nil, err
	}

	grpcClientInstance.Logger.Info("successfully obtained response from rpc server", zap.Any("Response", rpcResponse))

	return rpcResponse, err
}
