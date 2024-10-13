package middlewares

import (
	"bytes"
	"net/http"
)

// responseCapture captures the response data for later use
type responseCapture struct {
	http.ResponseWriter
	buffer *bytes.Buffer
	status int
}

func NewResponseCapture(rw http.ResponseWriter) *responseCapture {
	return &responseCapture{ResponseWriter: rw, buffer: bytes.NewBuffer(nil)}
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	// Capture the response into the buffer
	return rc.buffer.Write(b)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.status = statusCode
}

func (rc *responseCapture) Status() int {
	return rc.status
}

func (rc *responseCapture) Buffer() []byte {
	return rc.buffer.Bytes()
}

func (rc *responseCapture) Flush() (int, error) {
	rc.ResponseWriter.WriteHeader(rc.status)
	return rc.ResponseWriter.Write(rc.buffer.Bytes())
}
