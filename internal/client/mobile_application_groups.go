// Package client provides a client for interacting with the Link11 WAAP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListMobileApplicationGroups retrieves all mobile application groups in a configuration
func (c *Client) ListMobileApplicationGroups(ctx context.Context, configID string) ([]MobileApplicationGroup, error) {
	path := fmt.Sprintf("/conf/%s/mobile-application-groups", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[MobileApplicationGroup]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetMobileApplicationGroup retrieves a specific mobile application group
func (c *Client) GetMobileApplicationGroup(ctx context.Context, configID, entryID string) (*MobileApplicationGroup, error) {
	path := fmt.Sprintf("/conf/%s/mobile-application-groups/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result MobileApplicationGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateMobileApplicationGroup creates a new mobile application group
func (c *Client) CreateMobileApplicationGroup(ctx context.Context, configID, entryID string, mag *MobileApplicationGroup) error {
	path := fmt.Sprintf("/conf/%s/mobile-application-groups/%s", configID, entryID)

	body, err := json.Marshal(mag)
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

// UpdateMobileApplicationGroup updates an existing mobile application group
func (c *Client) UpdateMobileApplicationGroup(ctx context.Context, configID, entryID string, mag *MobileApplicationGroup) error {
	path := fmt.Sprintf("/conf/%s/mobile-application-groups/%s", configID, entryID)

	body, err := json.Marshal(mag)
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

// DeleteMobileApplicationGroup deletes a mobile application group
func (c *Client) DeleteMobileApplicationGroup(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/mobile-application-groups/%s", configID, entryID)
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
