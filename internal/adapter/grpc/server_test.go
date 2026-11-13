package grpc

import (
	"fmt"
	"testing"

	apperrors "ledger/internal/adapter/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToStatus(t *testing.T) {
	other := fmt.Errorf("boom")

	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"unauthenticated", apperrors.ErrUnauthenticated, codes.Unauthenticated},
		{"forbidden", fmt.Errorf("wrapped: %w", apperrors.ErrForbidden), codes.PermissionDenied},
		{"other errors pass through", other, codes.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := status.Code(toStatus(tt.err)); got != tt.want {
				t.Fatalf("code = %v, want %v", got, tt.want)
			}
		})
	}
}
