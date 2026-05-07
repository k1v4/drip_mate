package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k1v4/drip_mate/internal/entity"
)

func TestMLClient_GetRecommendation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/recommend" {
			t.Errorf("expected /recommend path, got %s", r.URL.Path)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"outfit": [
				{
					"item_id": "shirt-1",
					"score": 0.95,
					"category": "shirt",
					"material": "cotton"
				},
				{
					"item_id": "pants-1",
					"score": 0.88,
					"category": "pants",
					"material": "denim"
				}
			],
			"model_phase": "prod"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	req := &entity.RequestData{}

	result, err := client.GetRecommendation(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 outfit items, got %d", len(result))
	}

	if result[0].ItemID != "shirt-1" {
		t.Errorf("unexpected item id: %s", result[0].ItemID)
	}

	if result[0].Score != 0.95 {
		t.Errorf("unexpected score: %f", result[0].Score)
	}

	if result[0].Category != "shirt" {
		t.Errorf("unexpected category: %s", result[0].Category)
	}

	if result[0].Material != "cotton" {
		t.Errorf("unexpected material: %s", result[0].Material)
	}
}

func TestMLClient_GetRecommendation_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	req := &entity.RequestData{}

	_, err := client.GetRecommendation(context.Background(), req)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMLClient_GetRecommendation_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`invalid-json`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	req := &entity.RequestData{}

	_, err := client.GetRecommendation(context.Background(), req)

	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestMLClient_GetRecommendation_EmptyOutfit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{
			"outfit": [],
			"model_phase": "prod"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	req := &entity.RequestData{}

	result, err := client.GetRecommendation(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")

	if client == nil {
		t.Fatal("expected client")
	}

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("unexpected baseURL: %s", client.baseURL)
	}

	if client.httpClient == nil {
		t.Fatal("expected http client")
	}
}
