// Package client provides a client for interacting with the Link11 WAAP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListACLProfiles retrieves all ACL profiles in a configuration
func (c *Client) ListACLProfiles(ctx context.Context, configID string) ([]ACLProfile, error) {
	path := fmt.Sprintf("/conf/%s/acl-profiles", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[ACLProfile]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetACLProfile retrieves a specific ACL profile
func (c *Client) GetACLProfile(ctx context.Context, configID, entryID string) (*ACLProfile, error) {
	path := fmt.Sprintf("/conf/%s/acl-profiles/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ACLProfile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateACLProfile creates a new ACL profile
func (c *Client) CreateACLProfile(ctx context.Context, configID, entryID string, profile *ACLProfile) error {
	path := fmt.Sprintf("/conf/%s/acl-profiles/%s", configID, entryID)

	body, err := json.Marshal(profile)
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

// UpdateACLProfile updates an existing ACL profile
func (c *Client) UpdateACLProfile(ctx context.Context, configID, entryID string, profile *ACLProfile) error {
	path := fmt.Sprintf("/conf/%s/acl-profiles/%s", configID, entryID)

	body, err := json.Marshal(profile)
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

// DeleteACLProfile deletes an ACL profile
func (c *Client) DeleteACLProfile(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/acl-profiles/%s", configID, entryID)
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
