package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListDynamicRules retrieves all dynamic rules in a configuration
func (c *Client) ListDynamicRules(ctx context.Context, configID string) ([]DynamicRule, error) {
	path := fmt.Sprintf("/conf/%s/dynamic-rules", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[DynamicRule]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetDynamicRule retrieves a specific dynamic rule
func (c *Client) GetDynamicRule(ctx context.Context, configID, entryID string) (*DynamicRule, error) {
	path := fmt.Sprintf("/conf/%s/dynamic-rules/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result DynamicRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateDynamicRule creates a new dynamic rule
func (c *Client) CreateDynamicRule(ctx context.Context, configID, entryID string, rule *DynamicRule) error {
	path := fmt.Sprintf("/conf/%s/dynamic-rules/%s", configID, entryID)

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

// UpdateDynamicRule updates an existing dynamic rule
func (c *Client) UpdateDynamicRule(ctx context.Context, configID, entryID string, rule *DynamicRule) error {
	path := fmt.Sprintf("/conf/%s/dynamic-rules/%s", configID, entryID)

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

// DeleteDynamicRule deletes a dynamic rule
func (c *Client) DeleteDynamicRule(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/dynamic-rules/%s", configID, entryID)
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
