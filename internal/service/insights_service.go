package service

import (
	"context"
	"fmt"

	"github.com/natyb/itau-bff/internal/cache"
	"github.com/natyb/itau-bff/internal/clients"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
)

type ProfileClient interface {
	GetProfile(ctx context.Context, customerID string) (*clients.CustomerProfile, error)
}

type TransactionsClient interface {
	GetSummary(ctx context.Context, customerID string) (*clients.TransactionsSummary, error)
}

type InsightResult struct {
	CustomerID string
	Insight    string
	Cached     bool
}

type InsightsService struct {
	profileClient      ProfileClient
	transactionsClient TransactionsClient
	cache              *cache.Cache[*InsightResult]
}

func NewInsightsService(
	profileClient ProfileClient,
	transactionsClient TransactionsClient,
	cache *cache.Cache[*InsightResult],
) *InsightsService {
	return &InsightsService{
		profileClient:      profileClient,
		transactionsClient: transactionsClient,
		cache:              cache,
	}
}

func (s *InsightsService) GenerateInsight(ctx context.Context, customerID string) (*InsightResult, error) {
	tracer := otel.Tracer("insights-service")
	ctx, span := tracer.Start(ctx, "GenerateInsight")
	defer span.End()

	// Cache
	if cached, ok := s.cache.Get(customerID); ok {
		cached.Cached = true
		return cached, nil
	}

	g, ctx := errgroup.WithContext(ctx)

	var profile *clients.CustomerProfile
	var transactions *clients.TransactionsSummary

	g.Go(func() error {
		p, err := s.profileClient.GetProfile(ctx, customerID)
		if err != nil {
			return err
		}
		profile = p
		return nil
	})

	g.Go(func() error {
		t, err := s.transactionsClient.GetSummary(ctx, customerID)
		if err != nil {
			return err
		}
		transactions = t
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result := &InsightResult{
		CustomerID: customerID,
		Insight: fmt.Sprintf(
			"Cliente %s realizou %d transações recentemente, totalizando %.2f.",
			profile.Name,
			transactions.LastTransactions,
			transactions.TotalAmount,
		),
		Cached: false,
	}

	s.cache.Set(customerID, result)

	return result, nil
}
