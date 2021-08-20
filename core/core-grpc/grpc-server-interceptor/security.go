package grpc_server_interceptor

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAuthentication() grpc.UnaryServerInterceptor {
	return grpc_auth.UnaryServerInterceptor(securityContextHandle)
}

func StreamAuthentication() grpc.StreamServerInterceptor {
	return grpc_auth.StreamServerInterceptor(securityContextHandle)
}

func securityContextHandle(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}

	jwtToken, ok := md["jwt"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Authorization token is not supplied")
	}

	if jwtToken[0] != "" {
		return nil, status.Errorf(codes.Unauthenticated, "Authorization token is not supplied")
	}

	type authInfo struct {
		JwtToken string
	}

	newCtx := context.WithValue(ctx, "AuthToken", authInfo{jwtToken[0]})
	return newCtx, nil
}
