package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/natyb/itau-bff/internal/service"
	"log/slog"
	"os"
)

//
// Mock Service (implementa a interface esperada pelo handler)
//

type mockInsightsService struct {
	result *service.InsightResult
	err    error
}

func (m *mockInsightsService) GenerateInsight(ctx context.Context, customerID string) (*service.InsightResult, error) {
	return m.result, m.err
}

//
// Helpers
//

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

//
// Tests
//

func TestGetInsight_Success(t *testing.T) {
	mockService := &mockInsightsService{
		result: &service.InsightResult{
			CustomerID: "abc123",
			Insight:    "Insight de teste",
			Cached:     false,
		},
	}

	handler := NewInsightsHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/abc123", nil)
	req.SetPathValue("customerId", "abc123")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp InsightResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.CustomerID != "abc123" {
		t.Errorf("unexpected customerId")
	}

	if resp.Meta.Cached {
		t.Errorf("expected cached=false")
	}
}

func TestGetInsight_InvalidCustomerId(t *testing.T) {
	handler := NewInsightsHandler(&mockInsightsService{}, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/!!", nil)
	req.SetPathValue("customerId", "!!")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestGetInsight_Timeout(t *testing.T) {
	mockService := &mockInsightsService{
		err: context.DeadlineExceeded,
	}

	handler := NewInsightsHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/timeout123", nil)
	req.SetPathValue("customerId", "timeout123")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", rr.Code)
	}
}

func TestGetInsight_DependencyFailure(t *testing.T) {
	mockService := &mockInsightsService{
		err: errors.New("downstream error"),
	}

	handler := NewInsightsHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/insights/abc123", nil)
	req.SetPathValue("customerId", "abc123")

	rr := httptest.NewRecorder()

	handler.GetInsight(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rr.Code)
	}
}
