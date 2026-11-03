package oauth

import (
	"fmt"
	"strings"
)

type Token struct {
	Access_token	string	`json:"access_token"`
	Token_type		string	`json:"token_type"`
	Expires_in		int64	`json:"expires_in"`
	Scope			string	`json:"scope"`
}

func (t Token) Validate() error {
	if t.Access_token == "" {
		return fmt.Errorf("Oauth: Empty access token")
	}
	
	if !strings.EqualFold(t.Token_type, "Bearer") {
		return fmt.Errorf("Oauth: Unsupported token type: %q", t.Token_type)
	}
	
	if t.Expires_in < 0 {
		return fmt.Errorf("Oauth: Negative token lifetime")
	}
	
	return nil
}