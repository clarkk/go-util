package call

import (
	"fmt"
	"net/http"
)

type Error struct {
	status	int
	header	http.Header
	body	any
}

func (e *Error) Error() string {
	return fmt.Sprintf("HTTP %d: %+v", e.status, e.body)
}