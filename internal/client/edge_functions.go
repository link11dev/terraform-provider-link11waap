package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListEdgeFunctions retrieves all edge functions in a configuration
func (c *Client) ListEdgeFunctions(ctx context.Context, configID string) ([]EdgeFunction, error) {
	path := fmt.Sprintf("/conf/%s/edge-functions", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[EdgeFunction]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetEdgeFunction retrieves a specific edge function
func (c *Client) GetEdgeFunction(ctx context.Context, configID, entryID string) (*EdgeFunction, error) {
	path := fmt.Sprintf("/conf/%s/edge-functions/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result EdgeFunction
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateEdgeFunction creates a new edge function
func (c *Client) CreateEdgeFunction(ctx context.Context, configID, entryID string, ef *EdgeFunction) error {
	path := fmt.Sprintf("/conf/%s/edge-functions/%s", configID, entryID)

	body, err := json.Marshal(ef)
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

// UpdateEdgeFunction updates an existing edge function
func (c *Client) UpdateEdgeFunction(ctx context.Context, configID, entryID string, ef *EdgeFunction) error {
	path := fmt.Sprintf("/conf/%s/edge-functions/%s", configID, entryID)

	body, err := json.Marshal(ef)
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

// DeleteEdgeFunction deletes an edge function
func (c *Client) DeleteEdgeFunction(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/edge-functions/%s", configID, entryID)
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
