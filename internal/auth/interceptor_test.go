package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var secret = []byte("test-secret-that-is-at-least-32-bytes")

func token(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func TestUnaryInterceptor(t *testing.T) {
	valid := jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	expired := jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}
	noExpiry := jwt.RegisteredClaims{Subject: "user-1"}
	noSubject := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}

	tests := []struct {
		name    string
		header  string
		wantSub string
		wantErr bool
	}{
		{"valid token", "Bearer " + token(t, jwt.SigningMethodHS256, secret, valid), "user-1", false},
		{"no header", "", "", true},
		{"not a bearer token", token(t, jwt.SigningMethodHS256, secret, valid), "", true},
		{"wrong secret", "Bearer " + token(t, jwt.SigningMethodHS256, []byte("other"), valid), "", true},
		{"expired", "Bearer " + token(t, jwt.SigningMethodHS256, secret, expired), "", true},
		{"no expiry", "Bearer " + token(t, jwt.SigningMethodHS256, secret, noExpiry), "", true},
		{"no subject", "Bearer " + token(t, jwt.SigningMethodHS256, secret, noSubject), "", true},
		{"other algorithm", "Bearer " + token(t, jwt.SigningMethodHS512, secret, valid), "", true},
		{"unsigned", "Bearer " + token(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, valid), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.header != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", tt.header))
			}

			var gotSub string
			intercept, err := UnaryInterceptor(secret)
			if err != nil {
				t.Fatalf("UnaryInterceptor: %v", err)
			}

			_, err = intercept(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
				gotSub, _ = SubjectFromContext(ctx)
				return nil, nil
			})

			if tt.wantErr {
				if status.Code(err) != codes.Unauthenticated {
					t.Fatalf("code = %v, want Unauthenticated", status.Code(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if gotSub != tt.wantSub {
				t.Fatalf("subject = %q, want %q", gotSub, tt.wantSub)
			}
		})
	}
}

func TestUnaryInterceptorRejectsShortSecret(t *testing.T) {
	if _, err := UnaryInterceptor([]byte("too-short")); err == nil {
		t.Fatal("err = nil, want an error for a short secret")
	}
}
