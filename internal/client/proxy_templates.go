package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListProxyTemplates retrieves all proxy templates in a configuration
func (c *Client) ListProxyTemplates(ctx context.Context, configID string) ([]ProxyTemplate, error) {
	path := fmt.Sprintf("/conf/%s/proxy-templates", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[ProxyTemplate]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetProxyTemplate retrieves a specific proxy template
func (c *Client) GetProxyTemplate(ctx context.Context, configID, entryID string) (*ProxyTemplate, error) {
	path := fmt.Sprintf("/conf/%s/proxy-templates/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ProxyTemplate
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateProxyTemplate creates a new proxy template
func (c *Client) CreateProxyTemplate(ctx context.Context, configID, entryID string, pt *ProxyTemplate) error {
	path := fmt.Sprintf("/conf/%s/proxy-templates/%s", configID, entryID)

	body, err := json.Marshal(pt)
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

// UpdateProxyTemplate updates an existing proxy template
func (c *Client) UpdateProxyTemplate(ctx context.Context, configID, entryID string, pt *ProxyTemplate) error {
	path := fmt.Sprintf("/conf/%s/proxy-templates/%s", configID, entryID)

	body, err := json.Marshal(pt)
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

// DeleteProxyTemplate deletes a proxy template
func (c *Client) DeleteProxyTemplate(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/proxy-templates/%s", configID, entryID)
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
