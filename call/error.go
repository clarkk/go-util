package call

import (
	"fmt"
	"net/http"
)

type Error struct {
	url		string
	status	int
	header	http.Header
	body	any
}

func (e *Error) Error() string {
	return fmt.Sprintf("HTTP %d (%s): %+v", e.status, e.url, e.body)
}