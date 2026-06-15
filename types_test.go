package zabbix

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// ---------- ZBXBoolean ----------

func TestZBXBoolean_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"1"`, true},
		{`"true"`, true},
		{`"0"`, false},
		{`"false"`, false},
	}

	for _, tt := range tests {
		var b ZBXBoolean
		if err := json.Unmarshal([]byte(tt.input), &b); err != nil {
			t.Errorf("UnmarshalJSON(%s): unexpected error: %v", tt.input, err)
			continue
		}
		if bool(b) != tt.expected {
			t.Errorf("UnmarshalJSON(%s): expected %v, got %v", tt.input, tt.expected, bool(b))
		}
	}
}

func TestZBXBoolean_UnmarshalJSON_Invalid(t *testing.T) {
	var b ZBXBoolean
	err := json.Unmarshal([]byte(`"maybe"`), &b)
	if err == nil {
		t.Fatal("expected error for invalid boolean, got nil")
	}
	if !strings.Contains(err.Error(), "invalid input") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestZBXBoolean_MarshalJSON(t *testing.T) {
	tests := []struct {
		value    ZBXBoolean
		expected string
	}{
		{true, `"1"`},
		{false, `"0"`},
	}

	for _, tt := range tests {
		b, err := tt.value.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON(%v): unexpected error: %v", tt.value, err)
			continue
		}
		if string(b) != tt.expected {
			t.Errorf("MarshalJSON(%v): expected %s, got %s", tt.value, tt.expected, string(b))
		}
	}
}

func TestZBXBoolean_RoundTrip(t *testing.T) {
	original := ZBXBoolean(true)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ZBXBoolean
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip failed: expected %v, got %v", original, decoded)
	}
}

// ---------- UnixTimestamp ----------

func TestUnixTimestamp_UnmarshalJSON(t *testing.T) {
	var ts UnixTimestamp
	// 1609459200 = 2021-01-01 00:00:00 UTC
	if err := json.Unmarshal([]byte(`"1609459200"`), &ts); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if ts.Time == nil {
		t.Fatal("expected non-nil time after unmarshal")
	}
	expected := time.Unix(1609459200, 0).UTC()
	if ts.Unix() != expected.Unix() {
		t.Errorf("expected unix %d, got %d", expected.Unix(), ts.Unix())
	}
}

func TestUnixTimestamp_MarshalJSON(t *testing.T) {
	now := time.Unix(1609459200, 0)
	ts := UnixTimestamp{&now}
	b, err := ts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(b) != `"1609459200"` {
		t.Errorf("expected \"1609459200\", got %s", string(b))
	}
}

func TestUnixTimestamp_RoundTrip(t *testing.T) {
	original := time.Unix(1609459200, 0)
	ts := UnixTimestamp{&original}

	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded UnixTimestamp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.Unix() != original.Unix() {
		t.Errorf("round-trip mismatch: expected %d, got %d", original.Unix(), decoded.Unix())
	}
}

// ---------- Request ----------

func TestNewRequest_DefaultParams(t *testing.T) {
	req := NewRequest("host.get", nil)
	if req.JSONRPCVersion != "2.0" {
		t.Errorf("expected JSONRPCVersion '2.0', got %q", req.JSONRPCVersion)
	}
	if req.Method != "host.get" {
		t.Errorf("expected Method 'host.get', got %q", req.Method)
	}
	if req.RequestID == 0 {
		t.Error("expected non-zero RequestID")
	}
	if req.AuthToken != "" {
		t.Errorf("expected empty AuthToken, got %q", req.AuthToken)
	}
	// params should be non-nil (default empty map).
	if req.Params == nil {
		t.Error("expected non-nil Params for nil input")
	}
}

func TestNewRequest_WithParams(t *testing.T) {
	params := map[string]string{"hostid": "12345"}
	req := NewRequest("host.get", params)
	if req.Params == nil {
		t.Fatal("expected non-nil Params")
	}
}

func TestNewRequest_UniqueIDs(t *testing.T) {
	req1 := NewRequest("method1", nil)
	req2 := NewRequest("method2", nil)
	if req1.RequestID == req2.RequestID {
		t.Error("expected unique RequestIDs")
	}
}

func TestRequest_JSONSerialization(t *testing.T) {
	req := &Request{
		JSONRPCVersion: "2.0",
		Method:         "host.get",
		Params:         map[string]string{"hostid": "1"},
		RequestID:      42,
		AuthToken:      "token123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `"jsonrpc":"2.0"`) {
		t.Errorf("missing jsonrpc field: %s", s)
	}
	if !strings.Contains(s, `"method":"host.get"`) {
		t.Errorf("missing method field: %s", s)
	}
	if !strings.Contains(s, `"auth":"token123"`) {
		t.Errorf("missing auth field: %s", s)
	}
}

func TestRequest_OmitEmptyAuth(t *testing.T) {
	req := &Request{
		JSONRPCVersion: "2.0",
		Method:         "apiinfo.version",
		Params:         map[string]string{},
		RequestID:      1,
		AuthToken:      "",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// auth should be omitted when empty due to omitempty tag.
	if strings.Contains(string(data), `"auth"`) {
		t.Errorf("expected auth to be omitted, got: %s", string(data))
	}
}

// ---------- Response ----------

func TestResponse_Err_NoError(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Error:      APIError{Code: 0},
	}
	if err := resp.Err(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestResponse_Err_WithError(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Error: APIError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    "some detail",
		},
	}
	err := resp.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid params") {
		t.Errorf("error should contain message: %v", err)
	}
	if !strings.Contains(err.Error(), "-32602") {
		t.Errorf("error should contain code: %v", err)
	}
}

func TestResponse_Bind_Success(t *testing.T) {
	resp := &Response{
		Body: json.RawMessage(`[{"hostid":"1","host":"test"}]`),
	}
	var hosts []jHost
	if err := resp.Bind(&hosts); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if len(hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].HostID != "1" {
		t.Errorf("expected HostID '1', got %q", hosts[0].HostID)
	}
}

func TestResponse_Bind_InvalidJSON(t *testing.T) {
	resp := &Response{
		Body: json.RawMessage(`invalid json`),
	}
	var v interface{}
	if err := resp.Bind(&v); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ---------- APIError ----------

func TestAPIError_Error(t *testing.T) {
	e := APIError{Code: -32602, Message: "Invalid params"}
	errStr := e.Error()
	if !strings.Contains(errStr, "Invalid params") {
		t.Errorf("error should contain message: %s", errStr)
	}
	if !strings.Contains(errStr, "-32602") {
		t.Errorf("error should contain code: %s", errStr)
	}
}

func TestAPIError_AsError(t *testing.T) {
	e := &APIError{Code: -1, Message: "test"}
	// APIError should implement the error interface.
	var _ error = e
	if !errors.Is(e, e) {
		// APIError is a value type pointer; just verify it works as error.
	}
}
