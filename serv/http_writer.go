package serv

import (
	"net/http"
)

type Writer struct {
	http.ResponseWriter
	sent_headers bool
}

func (w *Writer) WriteHeader(code int){
	w.sent_headers = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *Writer) Write(b []byte) (int, error){
	w.sent_headers = true
	return w.ResponseWriter.Write(b)
}

func (w Writer) Sent_headers() bool {
	return w.sent_headers
}