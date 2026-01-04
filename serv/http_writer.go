package serv

import (
	"fmt"
	"bufio"
	"net"
	"net/http"
)

type Writer struct {
	http.ResponseWriter
	sent_header		bool
	status			int
	bytes_sent		int
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		ResponseWriter:	w,
		status:			http.StatusOK,
	}
}

func (w *Writer) WriteHeader(status int){
	if w.sent_header {
		return
	}
	w.sent_header	= true
	w.status		= status
	w.ResponseWriter.WriteHeader(status)
}

func (w *Writer) Write(b []byte) (int, error){
	if !w.sent_header {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes_sent += n
	return n, err
}

func (w *Writer) Status() int {
	return w.status
}

func (w *Writer) Sent() int {
	return w.bytes_sent
}

func (w *Writer) Sent_header() bool {
	return w.sent_header
}

//	Flush implements the http.Flusher interface.
//	This is critical for 2026 streaming APIs and Server-Sent Events (SSE).
func (w *Writer) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

//	Hijack implements the http.Hijacker interface for WebSockets.
func (w *Writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("Underlying ResponseWriter does not support hijacking")
}

//	Unwrap allows access to the original http.ResponseWriter (Go 1.20+ standard).
func (w *Writer) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}