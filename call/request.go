package call

import (
	"io"
	"fmt"
	"time"
	"mime"
	"bytes"
	"strings"
	"context"
	"net/http"
	"encoding/json/v2"
)

const (
	CONTENT_TYPE	= "Content-Type"
	TYPE_JSON		= "application/json"
	IDEMPOTENCY_KEY = "Idempotency-Key"
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
	user, pass, _ := strings.Cut(auth, ":")
	return func(req *http.Request){
		req.SetBasicAuth(user, pass)
	}
}

func Option_bearer_auth(auth string) Option {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
}

func Option_idempotency(key string) Option {
	return Option_header(IDEMPOTENCY_KEY, key)
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
	
	req.Header.Set("Accept", TYPE_JSON)
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
	res, err := c.http_client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	
	header		:= res.Header
	status		:= res.StatusCode
	
	if status == http.StatusNoContent {
		return status, header, nil
	}
	
	content_type	:= header.Get(CONTENT_TYPE)
	out_json		:= response_JSON(content_type)
	
	if status < 200 || status >= 300 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return status, header, fmt.Errorf("HTTP %d: Unable to read response body: %w", status, err)
		}
		
		if out_json {
			if out != nil {
				json.Unmarshal(b, out)
			}
			
			var out_err any
			if err := json.Unmarshal(b, &out_err); err == nil {
				return status, header, &Error{
					req.URL.String(),
					status,
					header,
					out_err,
				}
			}
		}
		return status, header, &Error{
			req.URL.String(),
			status,
			header,
			string(b),
		}
	}
	
	if out != nil {
		if !out_json {
			return status, header, &Error{
				req.URL.String(),
				status,
				header,
				fmt.Sprintf("Expected JSON response, but got: %s", content_type),
			}
		}
		if err := json.UnmarshalRead(res.Body, out); err != nil {
			return status, header, &Error{
				req.URL.String(),
				status,
				header,
				fmt.Sprintf("JSON unmarshal error: %v", err),
			}
		}
		return status, header, nil
	}
	
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return status, header, fmt.Errorf("HTTP %d: Unable to read response body: %w", status, err)
	}
	
	return status, header, nil
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

func response_JSON(content_type string) bool {
	media_type, _, err := mime.ParseMediaType(content_type)
	if err != nil {
		return false
	}
	return media_type == TYPE_JSON || strings.HasSuffix(media_type, "+json")
}