package common

import (
	"encoding/json"
	"net/http"
)

type HTTPError struct {
	Code    int
	Error   jsonError
	Message string `json:"message"`
}

type jsonError struct {
	Error error
}

func (err jsonError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Error.Error())
}

type HTTPResult interface {
	SetResult(result interface{})
}

type EmptyResultObject struct {
}

func (o *EmptyResultObject) SetResult(result interface{}) {

}

type HTTPResponseWrapper struct {
	Ok      bool       `json:"ok"`
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Result  HTTPResult `json:"result,omitempty"`
	Error   error      `json:"error,omitempty"`
}

type StatusWriter struct {
	http.ResponseWriter
	status int
	length int
}

// Status provides an easy way to retrieve the status code
func (w *StatusWriter) Status() int {
	return w.status
}

// Size provides an easy way to retrieve the response size in bytes
func (w *StatusWriter) Size() int {
	return w.length
}

// Header returns & satisfies the http.ResponseWriter interface
func (w *StatusWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader satisfies the http.ResponseWriter interface and
// allows us to cach the status code
func (w *StatusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write satisfies the http.ResponseWriter interface and
// captures data written, in bytes
func (w *StatusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

type HTTPLogEntry struct {
	Host            string  `json:"host"`
	RemoteAddr      string  `json:"remoteaddr"`
	Method          string  `json:"method"`
	RequestURI      string  `json:"requesturi"`
	Proto           string  `json:"proto"`
	Status          int     `json:"status"`
	ContentLen      int     `json:"contentlen"`
	UserAgent       string  `json:"useragent"`
	DurationSeconds float64 `json:"duration"`
}
