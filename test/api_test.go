package test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	service "github.com/MdSadiqMd/Exchange-Rate-Service/internal/services"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/cache"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/config"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/endpoint"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/external"
	"github.com/MdSadiqMd/Exchange-Rate-Service/pkg/transport"
	gokitlog "github.com/go-kit/log"
	"github.com/joho/godotenv"
)

func setupTestServer() *httptest.Server {
	stdlog := log.New(os.Stdout, "", log.LstdFlags)
	logger := gokitlog.NewNopLogger()

	_ = godotenv.Load()
	cfg, err := config.Load("../config.yaml")
	if err != nil {
		stdlog.Fatalf("failed to load config: %v", err)
	}

	cache := cache.NewMemoryCache(time.Duration(cfg.Cache.TTL) * time.Second)
	api := external.NewClient(cfg.ExternalAPI.BaseURL, cfg.ExternalAPI.APIKey, cfg.ExternalAPI.Timeout)

	conversionService := service.NewConversionService(logger, api, cache)
	conversionEndpoints := endpoint.MakeConversionEndpoints(conversionService)

	handler := transport.MakeHTTPHandler(conversionEndpoints, logger)
	return httptest.NewServer(handler)
}

func TestAPI_Convert(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/convert?from=INR&to=USD&amount=100")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	success, ok := result["success"].(bool)
	if !ok || !success {
		t.Error("expected success to be true")
	}
}
