package ledger

import (
	"context"
	"errors"
	"testing"
)

func TestNewCanceledContext(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:1/cashapp?sslmode=disable")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc, closeFn, err := New(ctx)
	if closeFn != nil {
		defer closeFn()
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected startup to respect cancellation, got %v", err)
	}
	if svc != nil || closeFn != nil {
		t.Fatal("failed startup must not return a service or a close function")
	}
}
