package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListUsers retrieves all users grouped by organization
func (c *Client) ListUsers(ctx context.Context) ([]UserOrganization, error) {
	resp, err := c.Get(ctx, "/accounts/users")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result []UserOrganization
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result, nil
}

// GetUser retrieves a specific user by ID
func (c *Client) GetUser(ctx context.Context, entryID string) (*User, error) {
	path := fmt.Sprintf("/accounts/%s", entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateUser creates a new user and returns the generated ID
func (c *Client) CreateUser(ctx context.Context, req *UserCreateRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.Post(ctx, "/accounts/users", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ParseErrorResponse(resp)
	}

	var result UserCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.ID, nil
}

// UpdateUser updates a user's details
func (c *Client) UpdateUser(ctx context.Context, entryID string, req *UserUpdateRequest) error {
	path := fmt.Sprintf("/accounts/%s", entryID)

	body, err := json.Marshal(req)
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

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, entryID string) error {
	path := fmt.Sprintf("/accounts/%s", entryID)
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
