package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/natyb/itau-bff/internal/cache"
	"github.com/natyb/itau-bff/internal/clients"
	"github.com/natyb/itau-bff/internal/http/handlers"
	"github.com/natyb/itau-bff/internal/observability"
	"github.com/natyb/itau-bff/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	mux := http.NewServeMux()

	// Observabilidade
	logger := observability.NewLogger()
	observability.RegisterMetrics()

	shutdownTracing := observability.InitTracing("itau-bff")
	defer func() {
	if err := shutdownTracing(context.Background()); err != nil {
		log.Printf("failed to shutdown tracing: %v", err)
	}
	}()

	// Clients
	profileClient := clients.NewProfileClient()
	transactionsClient := clients.NewTransactionsClient()

	// Cache (TTL = 60s, conforme requisito)
	insightsCache := cache.NewCache[*service.InsightResult](60 * time.Second)

	// Service (agora com cache)
	insightsService := service.NewInsightsService(
		profileClient,
		transactionsClient,
		insightsCache,
	)

	// Handler
	insightsHandler := handlers.NewInsightsHandler(
		insightsService,
		logger,
	)

	// Rotas
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc(
		"/v1/insights/{customerId}",
		insightsHandler.GetInsight,
	)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	// Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("API running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
