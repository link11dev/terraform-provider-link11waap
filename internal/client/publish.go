package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Publish publishes a configuration
func (c *Client) Publish(ctx context.Context, configID string, buckets []PublishBucket) error {
	path := fmt.Sprintf("/tools/publish/%s", configID)

	body, err := json.Marshal(buckets)
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
