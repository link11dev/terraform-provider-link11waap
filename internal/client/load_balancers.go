package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListLoadBalancers retrieves all load balancers in a configuration
func (c *Client) ListLoadBalancers(ctx context.Context, configID string) ([]LoadBalancer, error) {
	path := fmt.Sprintf("/conf/%s/load-balancers", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[LoadBalancer]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetLoadBalancerRegions retrieves region configuration for all load balancers
func (c *Client) GetLoadBalancerRegions(ctx context.Context, configID string) (*LoadBalancerRegions, error) {
	path := fmt.Sprintf("/conf/%s/load-balancers/regions", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result LoadBalancerRegions
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// UpdateLoadBalancerRegions updates region configuration for load balancers
func (c *Client) UpdateLoadBalancerRegions(ctx context.Context, configID string, req *LoadBalancerRegionsUpdateRequest) error {
	path := fmt.Sprintf("/conf/%s/load-balancers/regions", configID)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.Post(ctx, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	return nil
}

// AttachCertificateToLoadBalancer attaches a certificate to a load balancer
func (c *Client) AttachCertificateToLoadBalancer(ctx context.Context, configID, lbName, certID string, opts AttachCertificateOptions) error {
	path := fmt.Sprintf("/conf/%s/load-balancers/%s/certificates/%s", configID, lbName, certID)

	params := url.Values{}
	params.Set("provider", opts.Provider)
	params.Set("region", opts.Region)
	params.Set("listener", opts.Listener)
	params.Set("listener-port", strconv.Itoa(opts.ListenerPort))
	params.Set("default", strconv.FormatBool(opts.IsDefault))
	params.Set("elbv2", strconv.FormatBool(opts.ELBv2))
	path = path + "?" + params.Encode()

	resp, err := c.Put(ctx, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp)
	}

	return nil
}

// DetachCertificateFromLoadBalancer detaches a certificate from a load balancer
func (c *Client) DetachCertificateFromLoadBalancer(ctx context.Context, configID, lbName string, opts DetachCertificateOptions) error {
	path := fmt.Sprintf("/conf/%s/load-balancers/%s/certificates", configID, lbName)

	params := url.Values{}
	params.Set("provider", opts.Provider)
	params.Set("region", opts.Region)

	if opts.CertificateID != "" {
		params.Set("certificate-id", opts.CertificateID)
	}
	if opts.Listener != "" {
		params.Set("listener", opts.Listener)
	}
	if opts.ListenerPort != "" {
		params.Set("listener-port", opts.ListenerPort)
	}
	params.Set("elbv2", strconv.FormatBool(opts.ELBv2))

	path = path + "?" + params.Encode()
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
