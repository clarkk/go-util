package oauth

import (
	"fmt"
	"sync"
	"time"
	"strings"
	"context"
	"net/url"
	"github.com/clarkk/go-util/call"
)

const (
	Auth_basic Auth_method = iota
	Auth_body
)

type (
	Config struct {
		Token_url		string
		Client_id		string
		Client_secret	string
		Scope			string
		Auth_method		Auth_method
	}
	
	Client struct {
		config		Config
		http		*call.Client
		
		lock		sync.Mutex
		token		Token
		expires		time.Time
	}
	
	Token struct {
		Access_token	string	`json:"access_token"`
		Token_type		string	`json:"token_type"`
		Expires_in		int64	`json:"expires_in"`
		Scope			string	`json:"scope"`
	}
	
	Auth_method uint8
)

func NewClient(config Config, timeout int) *Client {
	return &Client{
		config:	config,
		http:	call.NewClient(config.Token_url, timeout),
	}
}

func (c *Client) Access_token(ctx context.Context) (string, error){
	c.lock.Lock()
	defer c.lock.Unlock()
	
	if c.token.Access_token != "" && time.Now().Before(c.expires) {
		return c.token.Access_token, nil
	}
	
	start := time.Now()
	
	token, err := c.client_credentials(ctx)
	if err != nil {
		return "", err
	}
	
	c.token = token
	
	lifetime := time.Duration(token.Expires_in) * time.Second
	margin := min(30 * time.Second, lifetime / 10)
	c.expires = start.Add(lifetime - margin)
	
	return token.Access_token, nil
}

func (c *Client) client_credentials(ctx context.Context) (Token, error){
	in := url.Values{
		"grant_type": {"client_credentials"},
	}
	
	if c.config.Scope != "" {
		in.Set("scope", c.config.Scope)
	}
	
	var opts []call.Option
	
	switch c.config.Auth_method {
	case Auth_basic:
		opts = append(opts, call.Option_basic_auth(
			url.QueryEscape(c.config.Client_id)+":"+url.QueryEscape(c.config.Client_secret),
		))
		
	case Auth_body:
		in.Set("client_id", c.config.Client_id)
		in.Set("client_secret", c.config.Client_secret)
		
	default:
		return Token{}, fmt.Errorf("Oauth: Unsupported auth method: %d", c.config.Auth_method)
	}
	
	var out Token
	
	if _, _, err := c.http.Send(ctx, "", in, &out, opts...); err != nil {
		return Token{}, fmt.Errorf("Oauth: Fetch access token: %w", err)
	}
	
	if out.Access_token == "" {
		return Token{}, fmt.Errorf("Oauth: Empty access token")
	}
	
	if !strings.EqualFold(out.Token_type, "Bearer") {
		return Token{}, fmt.Errorf("Oauth: Unsupported token type: %q", out.Token_type)
	}
	
	if out.Expires_in < 0 {
		return Token{}, fmt.Errorf("Oauth: Negative token lifetime")
	}
	
	return out, nil
}