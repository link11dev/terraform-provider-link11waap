package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListRateLimitRules retrieves all rate limit rules in a configuration
func (c *Client) ListRateLimitRules(ctx context.Context, configID string) ([]RateLimitRule, error) {
	path := fmt.Sprintf("/conf/%s/rate-limit-rules", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[RateLimitRule]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetRateLimitRule retrieves a specific rate limit rule
func (c *Client) GetRateLimitRule(ctx context.Context, configID, entryID string) (*RateLimitRule, error) {
	path := fmt.Sprintf("/conf/%s/rate-limit-rules/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result RateLimitRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateRateLimitRule creates a new rate limit rule
func (c *Client) CreateRateLimitRule(ctx context.Context, configID, entryID string, rule *RateLimitRule) error {
	path := fmt.Sprintf("/conf/%s/rate-limit-rules/%s", configID, entryID)

	body, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.Post(ctx, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return ParseErrorResponse(resp)
	}

	return nil
}

// UpdateRateLimitRule updates an existing rate limit rule
func (c *Client) UpdateRateLimitRule(ctx context.Context, configID, entryID string, rule *RateLimitRule) error {
	path := fmt.Sprintf("/conf/%s/rate-limit-rules/%s", configID, entryID)

	body, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.Put(ctx, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	return nil
}

// DeleteRateLimitRule deletes a rate limit rule
func (c *Client) DeleteRateLimitRule(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/rate-limit-rules/%s", configID, entryID)
	resp, err := c.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	return nil
}
