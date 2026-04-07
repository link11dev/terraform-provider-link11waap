package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListSecurityPolicies retrieves all security policies in a configuration
func (c *Client) ListSecurityPolicies(ctx context.Context, configID string) ([]SecurityPolicy, error) {
	path := fmt.Sprintf("/conf/%s/security-policies", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[SecurityPolicy]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetSecurityPolicy retrieves a specific security policy
func (c *Client) GetSecurityPolicy(ctx context.Context, configID, entryID string) (*SecurityPolicy, error) {
	path := fmt.Sprintf("/conf/%s/security-policies/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result SecurityPolicy
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateSecurityPolicy creates a new security policy
func (c *Client) CreateSecurityPolicy(ctx context.Context, configID, entryID string, sp *SecurityPolicy) error {
	path := fmt.Sprintf("/conf/%s/security-policies/%s", configID, entryID)

	body, err := json.Marshal(sp)
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

// UpdateSecurityPolicy updates an existing security policy
func (c *Client) UpdateSecurityPolicy(ctx context.Context, configID, entryID string, sp *SecurityPolicy) error {
	path := fmt.Sprintf("/conf/%s/security-policies/%s", configID, entryID)

	body, err := json.Marshal(sp)
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

// DeleteSecurityPolicy deletes a security policy
func (c *Client) DeleteSecurityPolicy(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/security-policies/%s", configID, entryID)
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
