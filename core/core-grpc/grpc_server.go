package core_grpc

import (
	"crypto/tls"

	"github.com/apssouza22/grpc-production-go/grpcutils"
	grpcserver "github.com/apssouza22/grpc-production-go/server"
	tlscert "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-tlsCert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	Logger *zap.Logger
	Server grpcserver.GrpcServer
	Address string
}

// NewGrpcService Initializes a new instance of a grpc service
func NewGrpcService(addr string, logger *zap.Logger, enableTls bool, cert *tls.Certificate) *GrpcServer {
	serverBuilder := grpcserver.GrpcServerBuilder{}

	addInterceptors(&serverBuilder)
	serverBuilder.EnableReflection(true)

	if enableTls {
		if cert != nil {
			serverBuilder.SetTlsCert(&tlscert.Cert)
		} else {
			serverBuilder.SetTlsCert(cert)
		}
	}

	s := serverBuilder.Build()

	return &GrpcServer{Address: addr, Logger: logger, Server: s}
}

// StartGrpcServer starts a grpc service
// usage:
//  s := NewGrpcService(logger *zap.Logger, enableTls bool, cert *tls.Certificate)
//  s.StartGrpcServer(addr string, fn func(server *grpc.Server))
func (grpcSrvInstance *GrpcServer) StartGrpcServer(fn func(server *grpc.Server)) {
	s := grpcSrvInstance.Server
	l := grpcSrvInstance.Logger
	addr := grpcSrvInstance.Address
	s.RegisterService(fn)

	err := s.Start(addr)
	if err != nil {
		l.Fatal(err.Error())
	}

	grpcSrvInstance.awaitTermination()
}

// Shuts down grpc server
func (grpcSrvInstance *GrpcServer) awaitTermination() {
	s := grpcSrvInstance.Server
	l := grpcSrvInstance.Logger

	s.AwaitTermination(func() {
		l.Info("shutting down grpc server")
	})
}

// addInterceptors adds default rpc interceptors to grpc service instance
func addInterceptors(s *grpcserver.GrpcServerBuilder) {
	s.SetUnaryInterceptors(grpcutils.GetDefaultUnaryServerInterceptors())
	s.SetStreamInterceptors(grpcutils.GetDefaultStreamServerInterceptors())
}
