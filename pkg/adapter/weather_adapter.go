package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/k1v4/drip_mate/internal/entity"
)

var (
	ErrInvalidAPIKey = errors.New("invalid api key")
	ErrCityNotFound  = errors.New("city not found")
	ErrRateLimited   = errors.New("rate limited")
)

type WeatherProvider interface {
	GetCurrentWeather(ctx context.Context, city string) (*entity.Weather, error)
}

type weatherResponse struct {
	Name string `json:"name"`

	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`

	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`

	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`

	Cod int `json:"cod"`
}

type OpenWeatherAdapter struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func NewOpenWeatherAdapter(apiKey string) *OpenWeatherAdapter {
	return &OpenWeatherAdapter{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: "https://api.openweathermap.org",
	}
}

func (a *OpenWeatherAdapter) GetCurrentWeather(ctx context.Context, city string) (*entity.Weather, error) {
	url := fmt.Sprintf(
		"%s/data/2.5/weather?q=%s&appid=%s&units=metric",
		a.baseURL,
		url.QueryEscape(city),
		a.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var dto weatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	// --- проверка cod из тела ---
	if dto.Cod != 200 {
		switch dto.Cod {
		case 401:
			return nil, ErrInvalidAPIKey
		case 404:
			return nil, ErrCityNotFound
		case 429:
			return nil, ErrRateLimited
		default:
			return nil, fmt.Errorf("api error: %d", dto.Cod)
		}
	}

	description := ""
	condition := ""
	if len(dto.Weather) > 0 {
		description = dto.Weather[0].Description
		condition = dto.Weather[0].Main
	}

	return &entity.Weather{
		City:        dto.Name,
		Temperature: dto.Main.Temp,
		FeelsLike:   dto.Main.FeelsLike,
		Description: description,
		Condition:   condition,
		Humidity:    dto.Main.Humidity,
		WindSpeed:   dto.Wind.Speed,
	}, nil
}
