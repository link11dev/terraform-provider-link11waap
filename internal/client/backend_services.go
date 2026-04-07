package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListBackendServices retrieves all backend services in a configuration
func (c *Client) ListBackendServices(ctx context.Context, configID string) ([]BackendService, error) {
	path := fmt.Sprintf("/conf/%s/backend-services", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[BackendService]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetBackendService retrieves a specific backend service
func (c *Client) GetBackendService(ctx context.Context, configID, entryID string) (*BackendService, error) {
	path := fmt.Sprintf("/conf/%s/backend-services/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result BackendService
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateBackendService creates a new backend service
func (c *Client) CreateBackendService(ctx context.Context, configID, entryID string, bs *BackendService) error {
	path := fmt.Sprintf("/conf/%s/backend-services/%s", configID, entryID)

	body, err := json.Marshal(bs)
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

// UpdateBackendService updates an existing backend service
func (c *Client) UpdateBackendService(ctx context.Context, configID, entryID string, bs *BackendService) error {
	path := fmt.Sprintf("/conf/%s/backend-services/%s", configID, entryID)

	body, err := json.Marshal(bs)
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

// DeleteBackendService deletes a backend service
func (c *Client) DeleteBackendService(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/backend-services/%s", configID, entryID)
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
