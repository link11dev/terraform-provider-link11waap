// Package client provides a client for interacting with the Link11 WAAP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetPlanet retrieves a specific planet entry by ID from the list endpoint.
// The API may return either a list wrapper ({"total":..,"items":[..]}) or a
// plain Planet object; both are handled here.
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Try list response first
	var listResult ListResponse[Planet]
	if err := json.Unmarshal(body, &listResult); err == nil && listResult.Items != nil {
		for _, p := range listResult.Items {
			if p.ID == entryID {
				return &p, nil
			}
		}
		return nil, &APIError{Code: 404, Message: fmt.Sprintf("planet entry %q not found in config %q", entryID, configID)}
	}

	// Try plain Planet object
	var planet Planet
	if err := json.Unmarshal(body, &planet); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if planet.ID == entryID {
		return &planet, nil
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
