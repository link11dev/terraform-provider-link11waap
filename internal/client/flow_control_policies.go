package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListFlowControlPolicies retrieves all flow control policies in a configuration
func (c *Client) ListFlowControlPolicies(ctx context.Context, configID string) ([]FlowControl, error) {
	path := fmt.Sprintf("/conf/%s/flow-control-policies", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[FlowControl]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if result.Total != len(result.Items) {
		return nil, fmt.Errorf("server reported %d flow control policies, but only %d were returned",
			result.Total, len(result.Items))
	}

	return result.Items, nil
}

// GetFlowControlPolicy retrieves a specific flow control policy
func (c *Client) GetFlowControlPolicy(ctx context.Context, configID, entryID string) (*FlowControl, error) {
	path := fmt.Sprintf("/conf/%s/flow-control-policies/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result FlowControl
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateFlowControlPolicy creates a new flow control policy
func (c *Client) CreateFlowControlPolicy(ctx context.Context, configID, entryID string, policy *FlowControl) error {
	path := fmt.Sprintf("/conf/%s/flow-control-policies/%s", configID, entryID)

	body, err := json.Marshal(policy)
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

// UpdateFlowControlPolicy updates an existing flow control policy
func (c *Client) UpdateFlowControlPolicy(ctx context.Context, configID, entryID string, policy *FlowControl) error {
	path := fmt.Sprintf("/conf/%s/flow-control-policies/%s", configID, entryID)

	body, err := json.Marshal(policy)
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

// DeleteFlowControlPolicy deletes a flow control policy
func (c *Client) DeleteFlowControlPolicy(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/flow-control-policies/%s", configID, entryID)
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
