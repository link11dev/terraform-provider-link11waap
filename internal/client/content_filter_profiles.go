package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListContentFilterProfiles retrieves all content filter profiles in a configuration
func (c *Client) ListContentFilterProfiles(ctx context.Context, configID string) ([]ContentFilterProfile, error) {
	path := fmt.Sprintf("/conf/%s/content-filter-profiles", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[ContentFilterProfile]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetContentFilterProfile retrieves a specific content filter profile
func (c *Client) GetContentFilterProfile(ctx context.Context, configID, entryID string) (*ContentFilterProfile, error) {
	path := fmt.Sprintf("/conf/%s/content-filter-profiles/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ContentFilterProfile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateContentFilterProfile creates a new content filter profile
func (c *Client) CreateContentFilterProfile(ctx context.Context, configID, entryID string, p *ContentFilterProfile) error {
	path := fmt.Sprintf("/conf/%s/content-filter-profiles/%s", configID, entryID)

	body, err := json.Marshal(p)
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

// UpdateContentFilterProfile updates an existing content filter profile
func (c *Client) UpdateContentFilterProfile(ctx context.Context, configID, entryID string, p *ContentFilterProfile) error {
	path := fmt.Sprintf("/conf/%s/content-filter-profiles/%s", configID, entryID)

	body, err := json.Marshal(p)
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

// DeleteContentFilterProfile deletes a content filter profile
func (c *Client) DeleteContentFilterProfile(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/content-filter-profiles/%s", configID, entryID)
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
