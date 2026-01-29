package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"github.com/natyb/itau-bff/internal/cache"
	"github.com/natyb/itau-bff/internal/clients"
)

// Mocks
type mockProfileClient struct{}

func (m *mockProfileClient) GetProfile(ctx context.Context, customerID string) (*clients.CustomerProfile, error) {
	return &clients.CustomerProfile{
		Name: "Cliente Teste",
	}, nil
}

type mockTransactionsClient struct {
	delay time.Duration
}

func (m *mockTransactionsClient) GetSummary(ctx context.Context, customerID string) (*clients.TransactionsSummary, error) {
	select {
	case <-time.After(m.delay):
		return &clients.TransactionsSummary{
			LastTransactions: 3,
			TotalAmount:      100.0,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Teste de sucesso
func TestGenerateInsight_Success(t *testing.T) {
	cache := cache.NewCache[*InsightResult](1 * time.Minute)

	service := NewInsightsService(
		&mockProfileClient{},
		&mockTransactionsClient{delay: 10 * time.Millisecond},
		cache,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err := service.GenerateInsight(ctx, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CustomerID != "abc123" {
		t.Errorf("expected customerId abc123")
	}

	if result.Cached {
		t.Errorf("expected cached=false on first call")
	}
}

//Teste de cache
func TestGenerateInsight_UsesCache(t *testing.T) {
	cache := cache.NewCache[*InsightResult](1 * time.Minute)

	service := NewInsightsService(
		&mockProfileClient{},
		&mockTransactionsClient{delay: 10 * time.Millisecond},
		cache,
	)

	ctx := context.Background()

	_, _ = service.GenerateInsight(ctx, "abc123")
	result, _ := service.GenerateInsight(ctx, "abc123")

	if !result.Cached {
		t.Errorf("expected cached=true on second call")
	}
}

// Teste de timeout
func TestGenerateInsight_Timeout(t *testing.T) {
	cache := cache.NewCache[*InsightResult](1 * time.Minute)

	service := NewInsightsService(
		&mockProfileClient{},
		&mockTransactionsClient{delay: 3 * time.Second},
		cache,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := service.GenerateInsight(ctx, "timeout123")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected timeout error")
	}
}
