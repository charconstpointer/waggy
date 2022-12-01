package v1_deprecated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/syke99/waggy/v1/header"
	"github.com/syke99/waggy/v1/internal/pkg/models"
	"github.com/syke99/waggy/v1/internal/pkg/resources"
	"io"
	"net/http"
	"os"
	"strings"
)

// ResponseWriter used for writing an HTTP Response
type ResponseWriter struct {
	status     resources.StatusCode
	Header     *header.Header
	writer     io.Writer
	buffer     *bytes.Buffer
	defErrResp defaultResponse
	defResp    defaultResponse
}

type defaultResponse struct {
	status int
	body   []byte
}

// Resp initializes a new ResponseWriter to be used to write HTTP Responses
func Resp(opts ...RouteOption) *ResponseWriter {
	h := header.Header{}

	rw := ResponseWriter{
		status:     0,
		Header:     &h,
		writer:     os.Stdout,
		buffer:     bytes.NewBuffer(make([]byte, 0)),
		defErrResp: defaultResponse{},
		defResp:    defaultResponse{},
	}

	for i, opt := range opts {
		switch i {
		case 0:
			rw.defResp.status = opt.status
			rw.defResp.body = opt.body
		case 1:
			rw.defErrResp.status = opt.status
			rw.defErrResp.body = opt.body
		}
	}

	return &rw
}

// WriteHeader writes the provided statusCode Header
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = resources.StatusCode(statusCode)
}

// Write composes a response and writes the response to the ResponseWriter's underlying io.Writer.
// If a call to WriteHeader has not been made before calling this method, Write will call WriteHeader
// with the StatusOK (200) HTTP status code
func (w *ResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(StatusOK)
	}

	if !w.Header.Has("Content-Type") {
		w.Header.Set("Content-Type", http.DetectContentType(body))
	}

	payload := w.buildResponse(body)

	return w.writer.Write(payload)
}

// WriteDefaultResponse sets the Content-Type header after determining the Content-Type with http.DetectContentType,
// and then calls w.writer.Write() on the set default response. Returns an error if no default response has been set
func (w *ResponseWriter) WriteDefaultResponse() (int, error) {
	if len(w.defResp.body) == 0 {
		return 0, resources.NoDefaultResponse
	}

	w.WriteHeader(w.defResp.status)

	if !w.Header.Has("Content-Type") {
		w.Header.Set("Content-Type", http.DetectContentType(w.defResp.body))
	}

	payload := w.buildResponse(w.defResp.body)

	return w.writer.Write(payload)
}

// WriteDefaultErrorResponse sets the Content-Type header after determining the Content-Type with http.DetectContentType,
// and then calls w.writer.Write() on the set default response. Returns an error if no default response has been set
func (w *ResponseWriter) WriteDefaultErrorResponse() (int, error) {
	if len(w.defErrResp.body) == 0 {
		return 0, resources.NoDefaultErrorResponse
	}

	w.WriteHeader(w.defErrResp.status)

	w.Header.Set("Content-Type", "application/problem+json")

	err := models.ErrReponse{
		Type:   os.Getenv(resources.FullURL.String()),
		Detail: string(w.defErrResp.body),
		Status: w.defErrResp.status,
	}

	errBytes, _ := json.Marshal(err)

	payload := w.buildResponse(errBytes)

	return w.writer.Write(payload)
}

// Error composes a response and writes an HTTP Error Response to the ResponseWriter's underlying io.Writer.
// It calls WriteHeader with the provided statusCode before composing the Error response
func (w *ResponseWriter) Error(statusCode int, error string) (int, error) {
	w.WriteHeader(statusCode)

	w.Header.Set("Content-Type", "application/problem+json")

	err := models.ErrReponse{
		Type:   os.Getenv(resources.FullURL.String()),
		Detail: error,
		Status: statusCode,
	}

	errBytes, _ := json.Marshal(err)

	payload := w.buildResponse(errBytes)

	return os.Stdout.Write(payload)
}

func (w *ResponseWriter) buildResponse(payload []byte) []byte {
	w.buffer.Write([]byte(fmt.Sprintf("%s %d %s\n", os.Getenv(resources.Scheme.String()), w.status, w.status.GetStatusCodeName())))

	headerLines := make([][]byte, 0)

	for k, v := range w.Header.Loop() {
		if k == "" {
			continue
		}

		if k == resources.ContentType.String() {
			w.buffer.Write([]byte(fmt.Sprintf("%s: %s\n", k, strings.Join(v, "; "))))
		}

		w.buffer.Write([]byte(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", "))))
	}

	for _, headerLine := range headerLines {
		w.buffer.Write(headerLine)
	}

	w.buffer.Write([]byte("\n"))

	w.buffer.Write(payload)

	response := w.buffer.Bytes()

	w.buffer.Reset()

	return response
}