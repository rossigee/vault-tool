package vault

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	addr   string
	client *http.Client
}

type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

type StatusResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

type InitRootTokenResponse struct {
	Nonce string `json:"nonce"`
	T     int    `json:"t"`
	N     int    `json:"n"`
}

type UpdateRootTokenResponse struct {
	Nonce        string `json:"nonce"`
	Progress     int    `json:"progress"`
	Required     int    `json:"required"`
	Complete     bool   `json:"complete"`
	EncodedToken string `json:"encoded_token"`
}

func NewClient(addr string) *Client {
	// Create client with TLS config that doesn't verify certs (for testing/emergency use)
	// In production, this should use proper certificate verification
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &Client{
		addr:   addr,
		client: client,
	}
}

func (c *Client) Unseal(ctx context.Context, key string) (*UnsealResponse, error) {
	payload := map[string]string{"key": key}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/sys/unseal", c.addr), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unseal request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result UnsealResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/sys/seal-status", c.addr), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("seal status request failed with status %d", resp.StatusCode)
	}

	var result StatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) InitRootTokenGeneration(ctx context.Context, ttl string) (*InitRootTokenResponse, error) {
	// No pgp_keys and no otp - unseal keys will be provided instead
	// Vault will return an encoded token
	payload := map[string]interface{}{}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/sys/generate-root/attempt", c.addr), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generate root token initialization failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data InitRootTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateRootTokenGeneration(ctx context.Context, nonce, key string) (*UpdateRootTokenResponse, error) {
	payload := map[string]string{
		"key":   key,
		"nonce": nonce,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/sys/generate-root/update", c.addr), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generate root token update failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data UpdateRootTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
