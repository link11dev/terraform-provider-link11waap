package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListCertificates retrieves all certificates in a configuration
func (c *Client) ListCertificates(ctx context.Context, configID string) ([]Certificate, error) {
	path := fmt.Sprintf("/conf/%s/certificates", configID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result ListResponse[Certificate]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Items, nil
}

// GetCertificate retrieves a specific certificate
func (c *Client) GetCertificate(ctx context.Context, configID, entryID string) (*Certificate, error) {
	path := fmt.Sprintf("/conf/%s/certificates/%s", configID, entryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ParseErrorResponse(resp)
	}

	var result Certificate
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// CreateCertificate creates a new certificate
// domains parameter is used for Let's Encrypt certificates
func (c *Client) CreateCertificate(ctx context.Context, configID, entryID string, cert *CertificateCreateRequest, domains []string) error {
	path := fmt.Sprintf("/conf/%s/certificates/%s", configID, entryID)

	// Add domains as query parameter if provided
	if len(domains) > 0 {
		params := url.Values{}
		for _, domain := range domains {
			params.Add("domains", domain)
		}
		path = path + "?" + params.Encode()
	}

	body, err := json.Marshal(cert)
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

// UpdateCertificate updates certificate settings (Let's Encrypt parameters)
func (c *Client) UpdateCertificate(ctx context.Context, configID, entryID string, leAutoRenew, leAutoReplace bool, replaceCertID string) error {
	path := fmt.Sprintf("/conf/%s/certificates/%s", configID, entryID)

	params := url.Values{}
	params.Set("le_auto_renew", fmt.Sprintf("%t", leAutoRenew))
	params.Set("le_auto_replace", fmt.Sprintf("%t", leAutoReplace))
	if replaceCertID != "" {
		params.Set("replace_cert_id", replaceCertID)
	}
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

// DeleteCertificate deletes a certificate
func (c *Client) DeleteCertificate(ctx context.Context, configID, entryID string) error {
	path := fmt.Sprintf("/conf/%s/certificates/%s", configID, entryID)
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
