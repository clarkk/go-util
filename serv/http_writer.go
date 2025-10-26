package serv

import "net/http"

type Writer struct {
	http.ResponseWriter
	sent_headers	bool
	status			int
	bytes_sent		int
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(status int){
	w.sent_headers	= true
	w.status		= status
	w.ResponseWriter.WriteHeader(status)
}

func (w *Writer) Write(b []byte) (int, error){
	if !w.sent_headers {
		w.sent_headers	= true
		w.status		= http.StatusOK
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

func (w *Writer) Sent_headers() bool {
	return w.sent_headers
}