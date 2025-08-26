package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	stdlog "log"

	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/config"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/endpoint"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/external"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/transport"
	kitlog "github.com/go-kit/log"
)

func main() {
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
	logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)

	_ = godotenv.Load()
	cfg, err := config.Load("config.yaml")
	if err != nil {
		stdlog.Fatalf("failed to load config: %v", err)
	}

	cache := cache.NewMemoryCache(time.Duration(cfg.Cache.TTL) * time.Second)
	api := external.NewClient(cfg.ExternalAPI.BaseURL, cfg.ExternalAPI.APIKey, cfg.ExternalAPI.Timeout)

	conversionService := service.NewConversionService(logger, api, cache)
	conversionEndpoints := endpoint.MakeConversionEndpoints(conversionService)

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	httpHandler := transport.MakeHTTPHandler(conversionEndpoints, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
	}

	go func() {
		stdlog.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			stdlog.Fatalf("server error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	stdlog.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		stdlog.Fatalf("server forced to shutdown: %s", err)
	}
	stdlog.Println("Server exited properly")
}
