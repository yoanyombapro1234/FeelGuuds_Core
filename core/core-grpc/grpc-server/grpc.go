package grpc_server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	core_logging "github.com/yoanyombapro1234/FeelGuuds/src/libraries/core/core-logging/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// GrpcServer GRPC server interface
type GrpcServer interface {
	Start(address string) error
	AwaitTermination(shutdownHook func())
	RegisterService(reg func(*grpc.Server))
	GetListener() net.Listener
}

// GrpcServerBuilder GRPC server builder
type GrpcServerBuilder struct {
	options                   []grpc.ServerOption
	enabledReflection         bool
	shutdownHook              func()
	enabledHealthCheck        bool
	disableDefaultHealthCheck bool
	logger                    core_logging.ILog
	serviceName               string
}

type grpcServer struct {
	server   *grpc.Server
	listener net.Listener
	logger   core_logging.ILog
}

func (s grpcServer) GetListener() net.Listener {
	return s.listener
}

// NewGrpcServerBuilder creates a new instance of gprc server builder object.
func NewGrpcServerBuilder(svcName string, logger core_logging.ILog) *GrpcServerBuilder {
	return &GrpcServerBuilder{logger: logger, serviceName: svcName}
}

// AddOption DialOption configures how we set up the connection.
func (sb *GrpcServerBuilder) AddOption(o grpc.ServerOption) {
	sb.options = append(sb.options, o)
}

// EnableReflection enables the reflection
// gRPC Server Reflection provides information about publicly-accessible gRPC services on a server,
// and assists clients at runtime to construct RPC requests and responses without precompiled service information.
// It is used by gRPC CLI, which can be used to introspect server protos and send/receive test RPCs.
// Warning! We should not have this enabled in production
func (sb *GrpcServerBuilder) EnableReflection(e bool) {
	sb.enabledReflection = e
}

// DisableDefaultHealthCheck disables the default health check service
// Warning! if you disable the default health check you must provide a custom health check service
func (sb *GrpcServerBuilder) DisableDefaultHealthCheck(e bool) {
	sb.disableDefaultHealthCheck = e
}

// SetServerParameters is used to set keepalive and max-age parameters on the server-side.
func (sb *GrpcServerBuilder) SetServerParameters(serverParams keepalive.ServerParameters) {
	keepAlive := grpc.KeepaliveParams(serverParams)
	sb.AddOption(keepAlive)
}

// SetStreamInterceptors set a list of interceptors to the Grpc server for stream connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (sb *GrpcServerBuilder) SetStreamInterceptors(interceptors []grpc.StreamServerInterceptor) {
	chain := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(interceptors...))
	sb.AddOption(chain)
}

// SetUnaryInterceptors set a list of interceptors to the Grpc server for unary connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (sb *GrpcServerBuilder) SetUnaryInterceptors(interceptors []grpc.UnaryServerInterceptor) {
	chain := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))
	sb.AddOption(chain)
}

// SetTlsCert sets credentials for server connections
func (sb *GrpcServerBuilder) SetTlsCert(cert *tls.Certificate) {
	sb.AddOption(grpc.Creds(credentials.NewServerTLSFromCert(cert)))
}

// Build is responsible for building a GRPC server
func (sb *GrpcServerBuilder) Build() GrpcServer {
	srv := grpc.NewServer(sb.options...)
	if !sb.disableDefaultHealthCheck {
		server := health.NewServer()
		grpc_health_v1.RegisterHealthServer(srv, server)
		server.SetServingStatus(sb.serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	}

	if sb.enabledReflection {
		reflection.Register(srv)
	}
	return &grpcServer{srv, nil, sb.logger}
}

// RegisterService register the services to the server
func (s grpcServer) RegisterService(reg func(*grpc.Server)) {
	reg(s.server)
}

// Start the GRPC server
func (s *grpcServer) Start(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)

	if err != nil {
		msg := fmt.Sprintf("Failed to listen: %v", err)
		return errors.New(msg)
	}
	go s.serv()

	s.logger.Info("gRPC Server started on %s ", addr)
	return nil
}

// AwaitTermination makes the program wait for the signal termination
// Valid signal termination (SIGINT, SIGTERM)
func (s *grpcServer) AwaitTermination(shutdownHook func()) {
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT, syscall.SIGTERM)
	<-interruptSignal
	s.cleanup()
	if shutdownHook != nil {
		shutdownHook()
	}
}

func (s *grpcServer) cleanup() {
	s.logger.Info("Stopping the server")
	s.server.GracefulStop()
	s.logger.Info("Closing the listener")
	s.listener.Close()
	s.logger.Info("End of Program")
}

func (s *grpcServer) serv() {
	if err := s.server.Serve(s.listener); err != nil {
		s.logger.Error(err, "failed to serve: %v")
	}
}
