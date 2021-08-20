package grpc_server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	grpc_utils "github.com/yoanyombapro1234/FeelGuuds_Core/core/grpc-utils"
	tlscert "github.com/yoanyombapro1234/FeelGuuds_Core/core/core-tlsCert"
)

func TestBuildGrpcServer(t *testing.T) {
	builder := &GrpcServerBuilder{}
	builder.SetTlsCert(&tlscert.Cert)
	builder.DisableDefaultHealthCheck(true)
	builder.EnableReflection(true)
	builder.SetStreamInterceptors(grpc_utils.GetDefaultStreamServerInterceptors(nil))
	builder.SetUnaryInterceptors(grpc_utils.GetDefaultUnaryServerInterceptors(nil))
	server := builder.Build()
	assert.NotNil(t, server)
}
