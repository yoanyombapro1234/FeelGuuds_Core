package server

import (
	"log"

	"github.com/apssouza22/grpc-production-go/grpcutils"
	core_auth_sdk "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-auth-sdk"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Configurations struct {
	StatsDConnectionAddr string
	Logger *zap.Logger
	Client *core_auth_sdk.Client
	ServiceName string
	Origins []string
	EnableDelayMiddleware bool
	EnableRandomErrorMiddleware bool
	MinDelay  int
	MaxDelay  int
	DelayUnit string
	Version string
}

type ServiceMiddlewares struct {
	AuthenticationMiddleware *AuthenticationMiddleware
	CorsMiddleware *CorsMiddleware
	LoggingMiddleware *LoggingMiddleware
	MetricsMiddleware *MetricsMiddleware
	RandomDelayMiddleware *RandomDelayMiddleware
	RandomErrMiddleware *RandomErrMiddleware
	VersionMiddleware *VersionMiddleware
}

// InitializeMiddleware initializes a middleware object ecompassing every middleware in this library
func InitializeMiddleware(c *Configurations) *ServiceMiddlewares{
	if c == nil {
		log.Fatalf("invalid input argument. configurations cannot be nil")
	}
	var serviceMw ServiceMiddlewares

	serviceMw.AuthenticationMiddleware = NewAuthenticationMiddleware(c.Logger, c.Client, c.ServiceName)
	serviceMw.CorsMiddleware = NewCorsMiddleware(c.Origins)
	serviceMw.LoggingMiddleware = NewLoggingMiddleware(c.Logger)
	serviceMw.MetricsMiddleware = NewMetricsMiddleware(c.StatsDConnectionAddr, c.Logger)
	serviceMw.VersionMiddleware = NewVersionMw(c.Version)

	if c.EnableRandomErrorMiddleware {
		serviceMw.RandomErrMiddleware = NewRandomErrMiddleware(c.Logger)
	}

	if c.EnableDelayMiddleware {
		serviceMw.RandomDelayMiddleware = NewRandomDelayMiddleware(c.MinDelay, c.MaxDelay, c.DelayUnit)
	}

	return &serviceMw
}

// StreamInterceptor returns a set of stream interceptors
func (m *ServiceMiddlewares) StreamInterceptor() []grpc.StreamServerInterceptor {
	streamInterceptors := grpcutils.GetDefaultStreamServerInterceptors()
	streamInterceptors = append(streamInterceptors,
								m.AuthenticationMiddleware.StreamInterceptor(),
								m.LoggingMiddleware.StreamInterceptor(),
								m.VersionMiddleware.StreamInterceptor())

	if m.RandomDelayMiddleware != nil {
		streamInterceptors = append(streamInterceptors, m.RandomDelayMiddleware.StreamInterceptor())
	}

	if m.RandomErrMiddleware != nil {
		streamInterceptors = append(streamInterceptors, m.RandomErrMiddleware.StreamInterceptor())
	}

	return streamInterceptors
}

// UnaryInterceptor returns a set of unary interceptors
func (m *ServiceMiddlewares) UnaryInterceptor() []grpc.UnaryServerInterceptor {
	unaryInterceptors := grpcutils.GetDefaultUnaryServerInterceptors()
	unaryInterceptors = append(unaryInterceptors,
		m.AuthenticationMiddleware.UnaryInterceptor(),
		m.LoggingMiddleware.UnaryInterceptor(),
		m.VersionMiddleware.UnaryInterceptor())

	if m.RandomDelayMiddleware != nil {
		unaryInterceptors = append(unaryInterceptors, m.RandomDelayMiddleware.UnaryInterceptor())
	}

	if m.RandomErrMiddleware != nil {
		unaryInterceptors = append(unaryInterceptors, m.RandomErrMiddleware.UnaryInterceptor())
	}

	return unaryInterceptors
}
