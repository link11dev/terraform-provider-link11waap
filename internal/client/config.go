package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListConfigs retrieves all configurations
func (c *Client) ListConfigs(ctx context.Context) ([]WAAPConfig, error) {
	resp, err := c.Get(ctx, "/conf/configs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[WAAPConfig]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetConfig retrieves a specific configuration
func (c *Client) GetConfig(ctx context.Context, configID string) (*WAAPConfig, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/conf/configs/%s", configID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result WAAPConfig
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}
