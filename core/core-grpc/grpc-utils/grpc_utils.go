package grpc_utils

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	grpc_client_interceptor "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-grpc/grpc-client-interceptor"
	grpc_server_interceptor "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-grpc/grpc-server-interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requestErrorHandler(p interface{}) (err error) {
	return status.Errorf(codes.Internal, "Something went wrong :( ")
}

// GetDefaultUnaryServerInterceptors returns the default interceptors server unary connections
func GetDefaultUnaryServerInterceptors(tracer opentracing.Tracer) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_server_interceptor.UnaryAuditServiceRequest(),
		grpc_server_interceptor.UnaryLogRequestCanceled(),
		grpc_server_interceptor.UnaryAuthentication(),
		grpc_opentracing.UnaryServerInterceptor(),
		otgrpc.OpenTracingServerInterceptor(tracer),
		// Recovery handlers should typically be last in the chain so that other middleware
		// (e.g. logging) can operate on the recovered state instead of being directly affected by any panic
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(requestErrorHandler)),
	}
}

// GetDefaultStreamServerInterceptors returns the default interceptors for server streams connections
func GetDefaultStreamServerInterceptors(tracer opentracing.Tracer) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_ctxtags.StreamServerInterceptor(),
		grpc_server_interceptor.StreamAuditServiceRequest(),
		grpc_server_interceptor.StreamLogRequestCanceled(),
		grpc_server_interceptor.StreamAuthentication(),
		otgrpc.OpenTracingStreamServerInterceptor(tracer),
		grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(requestErrorHandler)),
	}
}

// GetDefaultUnaryClientInterceptors returns the default interceptors for client unary connections
func GetDefaultUnaryClientInterceptors() []grpc.UnaryClientInterceptor {
	tracing := grpc_opentracing.UnaryClientInterceptor(
		grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
	)
	interceptors := []grpc.UnaryClientInterceptor{
		grpc_client_interceptor.UnaryTimeoutInterceptor(),
		tracing,
	}
	return interceptors
}

// GetDefaultStreamClientInterceptors returns the default interceptors for client stream connections
func GetDefaultStreamClientInterceptors() []grpc.StreamClientInterceptor {
	tracing := grpc_opentracing.StreamClientInterceptor(
		grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
	)
	interceptors := []grpc.StreamClientInterceptor{
		grpc_client_interceptor.StreamTimeoutInterceptor(),
		tracing,
	}
	return interceptors
}
