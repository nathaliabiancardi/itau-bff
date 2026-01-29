package clients

import (
	"context"
	"time"
	"go.opentelemetry.io/otel"
)

type TransactionsSummary struct {
	LastTransactions int
	TotalAmount      float64
}

type TransactionsClient struct{}

func NewTransactionsClient() *TransactionsClient {
	return &TransactionsClient{}
}

func (c *TransactionsClient) GetSummary(ctx context.Context, customerID string) (*TransactionsSummary, error) {
	tracer := otel.Tracer("transactions-client")
	ctx, span := tracer.Start(ctx, "GetSummary")
	defer span.End()

	select {
	case <-time.After(3 * time.Second):
		return &TransactionsSummary{
			LastTransactions: 5,
			TotalAmount:      1234.56,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
