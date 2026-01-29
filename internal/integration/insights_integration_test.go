package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"github.com/natyb/itau-bff/internal/cache"
	"github.com/natyb/itau-bff/internal/clients"
	"github.com/natyb/itau-bff/internal/http/handlers"
	"github.com/natyb/itau-bff/internal/service"
	"log/slog"
)

// Logger para testes
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// Client rápido (para cenário feliz)
type fastTransactionsClient struct{}

func (f *fastTransactionsClient) GetSummary(
	ctx context.Context,
	customerID string,
) (*clients.TransactionsSummary, error) {
	return &clients.TransactionsSummary{
		LastTransactions: 2,
		TotalAmount:      250.0,
	}, nil
}

// Teste de Integração — SUCCESS (200)
func TestInsightsIntegration_Success(t *testing.T) {
	profileClient := clients.NewProfileClient()
	transactionsClient := &fastTransactionsClient{}

	cache := cache.NewCache[*service.InsightResult](1 * time.Minute)

	insightsService := service.NewInsightsService(
		profileClient,
		transactionsClient,
		cache,
	)

	handler := handlers.NewInsightsHandler(insightsService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/abc123", nil)
	req.SetPathValue("customerId", "abc123")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

// Teste de Integração — TIMEOUT (504)
func TestInsightsIntegration_Timeout(t *testing.T) {
	// Clients reais (simulam latência)
	profileClient := clients.NewProfileClient()
	transactionsClient := clients.NewTransactionsClient()

	cache := cache.NewCache[*service.InsightResult](1 * time.Minute)

	insightsService := service.NewInsightsService(
		profileClient,
		transactionsClient,
		cache,
	)

	handler := handlers.NewInsightsHandler(insightsService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/timeout123", nil)
	req.SetPathValue("customerId", "timeout123")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", rr.Code)
	}
}
