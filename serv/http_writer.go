package serv

import "net/http"

type Writer struct {
	http.ResponseWriter
	sent_header		bool
	status			int
	bytes_sent		int
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(status int){
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