package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errUnauthenticated = status.Error(codes.Unauthenticated, "unauthenticated")

// MinSecretLength is the shortest HS256 secret accepted, so that tokens cannot be forged by brute force.
const MinSecretLength = 32

// UnaryInterceptor authenticates each call with an HS256 bearer token signed
// with secret and puts the token's subject in the context. It fails when the
// secret is shorter than MinSecretLength.
func UnaryInterceptor(secret []byte) (grpc.UnaryServerInterceptor, error) {
	if len(secret) < MinSecretLength {
		return nil, fmt.Errorf("auth secret must be at least %d bytes", MinSecretLength)
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, errUnauthenticated
		}

		token, found := strings.CutPrefix(values[0], "Bearer ")
		if !found {
			return nil, errUnauthenticated
		}

		claims := &jwt.RegisteredClaims{}
		if _, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired()); err != nil {
			return nil, errUnauthenticated
		}

		if claims.Subject == "" {
			return nil, errUnauthenticated
		}

		return handler(WithSubject(ctx, claims.Subject), req)
	}, nil
}
