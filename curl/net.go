package curl

import (
	"net/http"
	"crypto/tls"
)

//	Skip TLS cert verification on client HTTP requests (outbound)
func Skip_TLS_verify(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
}