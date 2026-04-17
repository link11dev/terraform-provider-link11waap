// Package client provides a client for interacting with the Link11 WAAP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListGlobalFilters retrieves all global filters in a configuration.
func (c *Client) ListGlobalFilters(ctx context.Context, configID string) ([]GlobalFilter, error) {
	path := fmt.Sprintf("/conf/%s/global-filters", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[GlobalFilter]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetGlobalFilter retrieves a specific global filter.
func (c *Client) GetGlobalFilter(ctx context.Context, configID, entryID string) (*GlobalFilter, error) {
	path := fmt.Sprintf("/conf/%s/global-filters/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result GlobalFilter
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateGlobalFilter creates a new global filter.
func (c *Client) CreateGlobalFilter(ctx context.Context, configID, entryID string, filter *GlobalFilter) error {
	path := fmt.Sprintf("/conf/%s/global-filters/%s", configID, entryID)

	body, err := json.Marshal(filter)
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

// UpdateGlobalFilter updates an existing global filter.
func (c *Client) UpdateGlobalFilter(ctx context.Context, configID, entryID string, filter *GlobalFilter) error {
	path := fmt.Sprintf("/conf/%s/global-filters/%s", configID, entryID)

	body, err := json.Marshal(filter)
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

// DeleteGlobalFilter deletes a global filter.
func (c *Client) DeleteGlobalFilter(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/global-filters/%s", configID, entryID)
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
