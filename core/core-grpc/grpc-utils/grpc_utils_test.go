package grpc_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultStreamClientInterceptors(t *testing.T) {
	clientInterceptors := GetDefaultStreamClientInterceptors()
	assert.NotEmpty(t, clientInterceptors)
}

func TestGetDefaultStreamServerInterceptors(t *testing.T) {
	serverInterceptors := GetDefaultStreamServerInterceptors(nil)
	assert.NotEmpty(t, serverInterceptors)
}

func TestGetDefaultUnaryClientInterceptors(t *testing.T) {
	clientInterceptors := GetDefaultUnaryClientInterceptors()
	assert.NotEmpty(t, clientInterceptors)
}

func TestGetDefaultUnaryServerInterceptors(t *testing.T) {
	serverInterceptors := GetDefaultUnaryServerInterceptors(nil)
	assert.NotEmpty(t, serverInterceptors)
}

func Test_requestErrorHandler(t *testing.T) {
	handler := requestErrorHandler("")
	assert.Error(t, handler)
}
