package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListActions retrieves all actions in a configuration
func (c *Client) ListActions(ctx context.Context, configID string) ([]Action, error) {
	path := fmt.Sprintf("/conf/%s/actions", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[Action]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetAction retrieves a specific action
func (c *Client) GetAction(ctx context.Context, configID, entryID string) (*Action, error) {
	path := fmt.Sprintf("/conf/%s/actions/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result Action
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateAction creates a new action.
func (c *Client) CreateAction(ctx context.Context, configID, entryID string, a *Action) error {
	path := fmt.Sprintf("/conf/%s/actions/%s", configID, entryID)

	body, err := json.Marshal(a)
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

// UpdateAction updates an existing action
func (c *Client) UpdateAction(ctx context.Context, configID, entryID string, a *Action) error {
	path := fmt.Sprintf("/conf/%s/actions/%s", configID, entryID)

	body, err := json.Marshal(a)
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

// DeleteAction deletes an action
func (c *Client) DeleteAction(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/actions/%s", configID, entryID)
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
