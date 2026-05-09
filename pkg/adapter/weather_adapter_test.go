package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenWeatherAdapter_GetCurrentWeather_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, _ = w.Write([]byte(`{
			"name": "Moscow",
			"cod": 200,
			"main": {
				"temp": 21.5,
				"feels_like": 20.0,
				"humidity": 65
			},
			"weather": [
				{
					"main": "Clouds",
					"description": "broken clouds"
				}
			],
			"wind": {
				"speed": 5.4
			}
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test-api-key")
	adapter.baseURL = server.URL

	weather, err := adapter.GetCurrentWeather(context.Background(), "Moscow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if weather.City != "Moscow" {
		t.Errorf("expected city Moscow, got %s", weather.City)
	}

	if weather.Temperature != 21.5 {
		t.Errorf("expected temp 21.5, got %f", weather.Temperature)
	}

	if weather.Description != "broken clouds" {
		t.Errorf("unexpected description: %s", weather.Description)
	}

	if weather.Condition != "Clouds" {
		t.Errorf("unexpected condition: %s", weather.Condition)
	}

	if weather.Humidity != 65 {
		t.Errorf("unexpected humidity: %d", weather.Humidity)
	}

	if weather.WindSpeed != 5.4 {
		t.Errorf("unexpected wind speed: %f", weather.WindSpeed)
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_InvalidAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"cod": 401
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("bad-key")
	adapter.baseURL = server.URL

	_, err := adapter.GetCurrentWeather(context.Background(), "Moscow")

	if err != ErrInvalidAPIKey {
		t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_CityNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"cod": 404
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test")
	adapter.baseURL = server.URL

	_, err := adapter.GetCurrentWeather(context.Background(), "UnknownCity")

	if err != ErrCityNotFound {
		t.Fatalf("expected ErrCityNotFound, got %v", err)
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"cod": 429
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test")
	adapter.baseURL = server.URL

	_, err := adapter.GetCurrentWeather(context.Background(), "Moscow")

	if err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_UnknownAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"cod": 500
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test")
	adapter.baseURL = server.URL

	_, err := adapter.GetCurrentWeather(context.Background(), "Moscow")

	if err == nil {
		t.Fatal("expected api error")
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`invalid-json`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test")
	adapter.baseURL = server.URL

	_, err := adapter.GetCurrentWeather(context.Background(), "Moscow")

	if err == nil {
		t.Fatal("expected json decode error")
	}
}

func TestOpenWeatherAdapter_GetCurrentWeather_EmptyWeatherArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"name": "Moscow",
			"cod": 200,
			"main": {
				"temp": 15,
				"feels_like": 13,
				"humidity": 70
			},
			"weather": [],
			"wind": {
				"speed": 3
			}
		}`))
	}))
	defer server.Close()

	adapter := NewOpenWeatherAdapter("test")
	adapter.baseURL = server.URL

	weather, err := adapter.GetCurrentWeather(context.Background(), "Moscow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if weather.Description != "" {
		t.Errorf("expected empty description, got %s", weather.Description)
	}

	if weather.Condition != "" {
		t.Errorf("expected empty condition, got %s", weather.Condition)
	}
}
