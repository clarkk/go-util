package curl

import (
	"io"
	"time"
	"net/http"
	"github.com/go-json-experiment/json"
)

func Curl_JSON(url string, input any, timeout int) (int, error){
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	r, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer r.Body.Close()
	
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return r.StatusCode, err
	}
	if err := json.Unmarshal(b, input); err != nil {
		return r.StatusCode, err
	}
	return r.StatusCode, nil
}