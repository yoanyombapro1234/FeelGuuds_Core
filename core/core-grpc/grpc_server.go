package core_grpc

import (
	"crypto/tls"
	"time"

	"github.com/apssouza22/grpc-production-go/grpcutils"
	grpcserver "github.com/apssouza22/grpc-production-go/server"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	tlscert "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-tlsCert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	addInterceptors(&serverBuilder, logger)
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
func addInterceptors(s *grpcserver.GrpcServerBuilder, logger *zap.Logger) {
	var grpcUnaryInterceptors []grpc.UnaryServerInterceptor = grpcutils.GetDefaultUnaryServerInterceptors()
	var grpcStreamInterceptors []grpc.StreamServerInterceptor = grpcutils.GetDefaultStreamServerInterceptors()

	opts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}

	grpcUnaryInterceptors = append(grpcUnaryInterceptors, grpc_zap.UnaryServerInterceptor(logger, opts...),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)))
	grpcStreamInterceptors = append(grpcStreamInterceptors, grpc_zap.StreamServerInterceptor(logger, opts...),
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)))

	s.SetUnaryInterceptors(grpcUnaryInterceptors)
	s.SetStreamInterceptors(grpcStreamInterceptors)
}
