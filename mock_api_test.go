package zabbix

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockAPI is a configurable JSON-RPC mock server used by unit tests.
// It routes requests by the "method" field in the JSON-RPC body and returns
// canned responses registered via handle().
type mockAPI struct {
	t        *testing.T
	server   *httptest.Server
	handlers map[string]func(req *Request) (result interface{}, err *APIError)

	// lastRequest captures the last received request for assertions.
	lastRequest *Request
	// lastAuthorization captures the last Authorization header value.
	lastAuthorization string
	// lastContentType captures the last Content-Type header value.
	lastContentType string
	// capturedBody holds the raw request body for inspection.
	capturedBody []byte
}

func newMockAPI(t *testing.T) *mockAPI {
	m := &mockAPI{
		t:        t,
		handlers: make(map[string]func(req *Request) (result interface{}, err *APIError)),
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handleHTTP))
	return m
}

func (m *mockAPI) handle(method string, fn func(req *Request) (result interface{}, err *APIError)) {
	m.handlers[method] = fn
}

func (m *mockAPI) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

func (m *mockAPI) URL() string {
	return m.server.URL + "/api_jsonrpc.php"
}

func (m *mockAPI) handleHTTP(w http.ResponseWriter, r *http.Request) {
	m.lastAuthorization = r.Header.Get("Authorization")
	m.lastContentType = r.Header.Get("Content-Type")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.capturedBody = body

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.lastRequest = &req

	w.Header().Set("Content-Type", "application/json")

	fn, ok := m.handlers[req.Method]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found in mock",
				"data":    req.Method,
			},
			"id": req.RequestID,
		})
		return
	}

	result, apiErr := fn(&req)
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      req.RequestID,
	}
	if apiErr != nil {
		resp["error"] = apiErr
	} else {
		resp["result"] = result
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		m.t.Fatalf("failed to encode mock response: %v", err)
	}
}

// rpcResult builds a JSON-RPC result envelope string for a raw JSON payload.
func rpcResult(id uint64, rawJSON string) string {
	return strings.Join([]string{
		`{"jsonrpc":"2.0","id":`, jsonIntString(id), `,"result":`,
		rawJSON, `}`,
	}, "")
}

func jsonIntString(n uint64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
