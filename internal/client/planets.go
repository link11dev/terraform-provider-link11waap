// Package client provides a client for interacting with the Link11 WAAP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPlanet retrieves a specific planet entry by ID from the list endpoint.
// The API only provides a list endpoint; this function filters by entry ID.
func (c *Client) GetPlanet(ctx context.Context, configID, entryID string) (*Planet, error) {
	path := fmt.Sprintf("/conf/%s/planets", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[Planet]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	for _, p := range result.Items {
		if p.ID == entryID {
			return &p, nil
		}
	}

	return nil, &APIError{Code: 404, Message: fmt.Sprintf("planet entry %q not found in config %q", entryID, configID)}
}

// UpsertPlanet creates or replaces a planet entry via PUT.
func (c *Client) UpsertPlanet(ctx context.Context, configID, entryID string, planet *Planet) error {
	path := fmt.Sprintf("/conf/%s/planets/%s", configID, entryID)

	body, err := json.Marshal(planet)
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
