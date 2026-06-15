package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListContentFilterRules retrieves all content filter rules in a configuration
func (c *Client) ListContentFilterRules(ctx context.Context, configID string) ([]ContentFilterRule, error) {
	path := fmt.Sprintf("/conf/%s/content-filter-rules", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[ContentFilterRule]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetContentFilterRule retrieves a specific content filter rule
func (c *Client) GetContentFilterRule(ctx context.Context, configID, entryID string) (*ContentFilterRule, error) {
	path := fmt.Sprintf("/conf/%s/content-filter-rules/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ContentFilterRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateContentFilterRule creates a new content filter rule
func (c *Client) CreateContentFilterRule(ctx context.Context, configID, entryID string, rule *ContentFilterRule) error {
	path := fmt.Sprintf("/conf/%s/content-filter-rules/%s", configID, entryID)

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

// UpdateContentFilterRule updates an existing content filter rule
func (c *Client) UpdateContentFilterRule(ctx context.Context, configID, entryID string, rule *ContentFilterRule) error {
	path := fmt.Sprintf("/conf/%s/content-filter-rules/%s", configID, entryID)

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

// DeleteContentFilterRule deletes a content filter rule
func (c *Client) DeleteContentFilterRule(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/content-filter-rules/%s", configID, entryID)
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
