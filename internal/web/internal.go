package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
)

// InternalRequest dispatches method+path through the same ServeMux used by the
// live HTTP servers, bypassing network I/O, authentication, and the mud lock.
//
// The caller is responsible for holding the mud lock when the handler requires
// it (i.e. when calling from outside the normal game loop). Handlers wrapped
// with RunWithMUDLocked detect internal requests via IsInternalRequest and skip
// re-acquiring the lock.
//
// body may be nil for requests that have no payload.
// The returned responseBody is the raw response bytes.
func InternalRequest(method, path string, body io.Reader) (statusCode int, responseBody []byte, err error) {
	r := httptest.NewRequest(method, path, body)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	r = r.WithContext(withInternalContext(r.Context()))

	w := httptest.NewRecorder()
	internalMux.ServeHTTP(w, r)

	result := w.Result()
	defer result.Body.Close()

	responseBody, err = io.ReadAll(result.Body)
	return result.StatusCode, responseBody, err
}

// InternalRequestJSON is a convenience wrapper around InternalRequest that
// marshals reqBody as JSON (pass nil for no body) and unmarshals the response
// into dst (pass nil to discard the response body).
func InternalRequestJSON(method, path string, reqBody any, dst any) (int, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	status, respBody, err := InternalRequest(method, path, bodyReader)
	if err != nil {
		return status, err
	}

	if dst != nil {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return status, fmt.Errorf("unmarshal response body: %w", err)
		}
	}

	return status, nil
}
