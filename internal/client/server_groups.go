package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListServerGroups retrieves all server groups in a configuration
func (c *Client) ListServerGroups(ctx context.Context, configID string) ([]ServerGroup, error) {
	path := fmt.Sprintf("/conf/%s/server-groups", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[ServerGroup]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetServerGroup retrieves a specific server group
func (c *Client) GetServerGroup(ctx context.Context, configID, entryID string) (*ServerGroup, error) {
	path := fmt.Sprintf("/conf/%s/server-groups/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ServerGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateServerGroup creates a new server group
func (c *Client) CreateServerGroup(ctx context.Context, configID, entryID string, sg *ServerGroupCreateRequest) error {
	path := fmt.Sprintf("/conf/%s/server-groups/%s", configID, entryID)

	body, err := json.Marshal(sg)
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

// UpdateServerGroup updates an existing server group
func (c *Client) UpdateServerGroup(ctx context.Context, configID, entryID string, sg *ServerGroupCreateRequest) error {
	path := fmt.Sprintf("/conf/%s/server-groups/%s", configID, entryID)

	body, err := json.Marshal(sg)
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

// DeleteServerGroup deletes a server group
func (c *Client) DeleteServerGroup(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/server-groups/%s", configID, entryID)
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
