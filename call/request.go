package call

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"strings"
	"context"
	"net/http"
	"encoding/json/v2"
)

const (
	CONTENT_TYPE	= "Content-Type"
	TYPE_JSON		= "application/json"
)

type (
	Client struct {
		endpoint	string
		http_client	*http.Client
		options		[]Option
	}
	
	Option func(*http.Request)
)

func Option_header(key, value string) Option {
	return func(req *http.Request){
		req.Header.Set(key, value)
	}
}

func Option_basic_auth(auth string) Option {
	return func(req *http.Request) {
		user, pass, _ := strings.Cut(auth, ":")
		req.SetBasicAuth(user, pass)
	}
}

func Option_idempotency(key string) Option {
	return Option_header("Idempotency-Key", key)
}

func NewClient(endpoint string, timeout int, opts ...Option) *Client {
	return &Client{
		endpoint:	endpoint,
		http_client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		options:	opts,
	}
}

func (c *Client) Send(ctx context.Context, uri string, in, out any, opts ...Option) (int, http.Header, error){
	method, body, err := c.payload(in)
	if err != nil {
		return 0, nil, err
	}
	
	url := "https://"+c.endpoint+uri
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	
	if in != nil {
		req.Header.Set(CONTENT_TYPE, TYPE_JSON)
	}
	
	for _, opt := range c.options {
		opt(req)
	}
	
	for _, opt := range opts {
		opt(req)
	}
	
	return c.request(req, out)
}

func (c *Client) request(req *http.Request, out any) (int, http.Header, error){
	resp, err := c.http_client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	
	header := resp.Header
	
	content_type	:= header.Get(CONTENT_TYPE)
	out_json		:= strings.HasPrefix(content_type, TYPE_JSON)
	
	if resp.StatusCode >= 400 {
		out_bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, header, fmt.Errorf("HTTP error (%d): Unable to read response body: %v", resp.StatusCode, err)
		}
		
		if out_json {
			var out_err map[string]any
			if err := json.UnmarshalRead(bytes.NewReader(out_bytes), &out_err); err == nil {
				return resp.StatusCode, header, fmt.Errorf("HTTP error (%d): %v", resp.StatusCode, out_err)
			}
		}
		return resp.StatusCode, header, fmt.Errorf("HTTP error (%d): %s", resp.StatusCode, string(out_bytes))
	}
	
	if out != nil {
		if !out_json {
			return resp.StatusCode, header, fmt.Errorf("Expected JSON response, but got: %s", content_type)
		}
		return resp.StatusCode, header, json.UnmarshalRead(resp.Body, out)
	}
	
	return resp.StatusCode, header, nil
}

func (c *Client) payload(in any) (string, io.Reader, error){
	if in == nil {
		return http.MethodGet, nil, nil
	}
	var buf bytes.Buffer
	if err := json.MarshalWrite(&buf, in); err != nil {
		return "", nil, err
	}
	return http.MethodPost, &buf, nil
}