package serv

import "net/http"

type Writer struct {
	http.ResponseWriter
	sent_headers	bool
	status			int
	bytes_sent		int
}

func (w *Writer) WriteHeader(status int){
	w.sent_headers	= true
	w.status		= status
	w.ResponseWriter.WriteHeader(status)
}

func (w *Writer) Write(b []byte) (int, error){
	n, err := w.ResponseWriter.Write(b)
	w.bytes_sent += n
	return n, err
}

func (w Writer) Sent_headers() bool {
	return w.sent_headers
}