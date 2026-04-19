package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/k1v4/drip_mate/internal/entity"
)

type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type outfitItem struct {
	ItemID   string  `json:"item_id"`
	Score    float64 `json:"score"`
	Category string  `json:"category"`
	Material string  `json:"material"`
}

type recommendResponse struct {
	Outfit     []outfitItem `json:"outfit"`
	ModelPhase string       `json:"model_phase"`
}

func (c *MLClient) GetRecommendation(ctx context.Context, data *entity.RequestData) ([]entity.OutfitItem, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("ml_client: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/recommend",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("ml_client: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml_client: do request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml_client: unexpected status %d", resp.StatusCode)
	}

	var result recommendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml_client: decode response: %w", err)
	}

	items := make([]entity.OutfitItem, len(result.Outfit))
	for i, item := range result.Outfit {
		items[i] = entity.OutfitItem{
			ItemID:   item.ItemID,
			Score:    item.Score,
			Category: item.Category,
			Material: item.Material,
		}
	}

	return items, nil
}
