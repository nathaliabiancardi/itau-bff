package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/natyb/itau-bff/internal/clients"
	"github.com/natyb/itau-bff/internal/observability"
	"github.com/natyb/itau-bff/internal/service"
	"go.opentelemetry.io/otel"
)

var customerIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

type InsightsService interface {
	GenerateInsight(ctx context.Context, customerID string) (*service.InsightResult, error)
}

type InsightsHandler struct {
	service InsightsService
	logger  *slog.Logger
}

func NewInsightsHandler(
	service InsightsService,
	logger *slog.Logger,
) *InsightsHandler {
	return &InsightsHandler{
		service: service,
		logger:  logger,
	}
}

type InsightResponse struct {
	CustomerID string `json:"customerId"`
	Insight    string `json:"insight"`
	Meta       Meta   `json:"meta"`
}

type Meta struct {
	Cached bool `json:"cached"`
}

func (h *InsightsHandler) GetInsight(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := r.Header.Get("X-Request-Id")
	customerID := r.PathValue("customerId")

	tracer := otel.Tracer("http-handler")
	ctx, span := tracer.Start(r.Context(), "GET /v1/insights")
	defer span.End()

	// Validação
	if !customerIDRegex.MatchString(customerID) {
		h.logger.Warn("invalid customerId",
			"customerId", customerID,
			"requestId", requestID,
		)

		http.Error(w, "invalid customerId", http.StatusBadRequest)
		return
	}

	// Timeout do BFF
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := h.service.GenerateInsight(ctxTimeout, customerID)
	if err != nil {
		status := http.StatusBadGateway
		message := "failed to generate insight"

		switch {
		case errors.Is(err, clients.ErrRateLimited):
			status = http.StatusTooManyRequests
			message = "too many requests"

		case errors.Is(err, clients.ErrCircuitOpen):
			status = http.StatusBadGateway
			message = "dependency unavailable"

		case errors.Is(err, context.DeadlineExceeded),
			errors.Is(err, context.Canceled):
			status = http.StatusGatewayTimeout
			message = "request timeout"
		}

		h.logger.Error("insight request failed",
			"customerId", customerID,
			"error", err.Error(),
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"requestId", requestID,
		)

		http.Error(w, message, status)
		return
	}

	response := InsightResponse{
		CustomerID: result.CustomerID,
		Insight:    result.Insight,
		Meta: Meta{
			Cached: result.Cached,
		},
	}

	h.logger.Info("insight generated",
		"customerId", customerID,
		"cached", result.Cached,
		"latency_ms", time.Since(start).Milliseconds(),
		"requestId", requestID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)

	observability.RequestDuration.WithLabelValues(
		"/v1/insights",
	).Observe(time.Since(start).Seconds())
}
