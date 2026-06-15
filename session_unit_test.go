package zabbix

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// helper to create a session against the mock with pre-set version.
func newMockSession(t *testing.T, version string) (*mockAPI, *Session) {
	t.Helper()
	m := newMockAPI(t)

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return version, nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return "test-token-123", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	return m, session
}

func TestNewSession_LoginSuccess(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	if session.Token != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got %q", session.Token)
	}
	if session.APIVersion != "6.0.0" {
		t.Errorf("expected version '6.0.0', got %q", session.APIVersion)
	}
	if session.ApiVersion.Major != 6 {
		t.Errorf("expected Major=6, got %d", session.ApiVersion.Major)
	}
	if session.ApiVersion.Minor != 0 {
		t.Errorf("expected Minor=0, got %d", session.ApiVersion.Minor)
	}
}

func TestNewSession_WithToken(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.4.15", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "", "", "preset-token")
	if err != nil {
		t.Fatalf("NewSession with token failed: %v", err)
	}
	if session.Token != "preset-token" {
		t.Errorf("expected token 'preset-token', got %q", session.Token)
	}
}

func TestNewSession_LoginError(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return nil, &APIError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    "Login name or password is incorrect.",
		}
	})

	_, err := NewSession(context.Background(), m.URL(), "wrong", "creds", "")
	if err == nil {
		t.Fatal("expected login error, got nil")
	}
	if !strings.Contains(err.Error(), "Error logging in") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetVersion_ParsesVersion(t *testing.T) {
	m, session := newMockSession(t, "7.0.5")
	defer m.Close()

	v, err := session.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	if v != "7.0.5" {
		t.Errorf("expected '7.0.5', got %q", v)
	}
	if session.ApiVersion.Major != 7 {
		t.Errorf("expected Major=7, got %d", session.ApiVersion.Major)
	}
	if session.ApiVersion.Minor != 0 {
		t.Errorf("expected Minor=0, got %d", session.ApiVersion.Minor)
	}
	if session.ApiVersion.Build != 5 {
		t.Errorf("expected Build=5, got %d", session.ApiVersion.Build)
	}
}

func TestGetVersion_Cached(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	callCount := 0
	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		callCount++
		return "6.0.0", nil
	})

	// First call was during login; second should be cached.
	_, _ = session.GetVersion(context.Background())
	_, _ = session.GetVersion(context.Background())

	// apiinfo.version should only have been called once (during login).
	if callCount != 0 {
		t.Errorf("expected 0 extra apiinfo calls (cached), got %d", callCount)
	}
}

func TestDo_BearerAuth_ForZabbix7(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "7.0.0", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "bearer-test-token")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, _ = session.GetHosts(context.Background(), HostGetParams{})

	if m.lastAuthorization != "Bearer bearer-test-token" {
		t.Errorf("expected Bearer header, got %q", m.lastAuthorization)
	}

	// Verify auth field is NOT in the JSON-RPC body for Zabbix 7+
	if m.lastRequest.AuthToken != "" {
		t.Errorf("expected empty auth in body for Zabbix 7+, got %q", m.lastRequest.AuthToken)
	}
}

func TestDo_LegacyAuth_ForZabbix6(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "legacy-token")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, _ = session.GetHosts(context.Background(), HostGetParams{})

	if m.lastAuthorization != "" {
		t.Errorf("expected no Authorization header for Zabbix 6, got %q", m.lastAuthorization)
	}
	if m.lastRequest.AuthToken != "legacy-token" {
		t.Errorf("expected auth in body for Zabbix 6, got %q", m.lastRequest.AuthToken)
	}
}

func TestDo_ContentType_JSON(t *testing.T) {
	m, session := newMockSession(t, "6.0.0")
	defer m.Close()

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, _ = session.GetHosts(context.Background(), HostGetParams{})

	if m.lastContentType != "application/json" {
		t.Errorf("expected 'application/json', got %q", m.lastContentType)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	session := &Session{URL: m.URL(), Token: "token"}
	session.ApiVersion.Major = 6

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := session.Do(ctx, NewRequest("host.get", nil))
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
}

func TestDo_APIError(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "token")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return nil, &APIError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    "Host with the same name already exists.",
		}
	})

	_, err = session.GetHosts(context.Background(), HostGetParams{})
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid params") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDo_EmptyResult(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	session, err := NewSession(context.Background(), m.URL(), "Admin", "zabbix", "token")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, err = session.GetHosts(context.Background(), HostGetParams{})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for empty result, got %v", err)
	}
}

func TestDo_GetVersionParseError(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "bad", nil
	})

	session := &Session{URL: m.URL()}
	// Should not panic even with malformed version "bad".
	v, err := session.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "bad" {
		t.Errorf("expected raw 'bad', got %q", v)
	}
	// Major should be 0 since "bad" can't be parsed.
	if session.ApiVersion.Major != 0 {
		t.Errorf("expected Major=0 for invalid version, got %d", session.ApiVersion.Major)
	}
}

func TestAuthToken_SetToken(t *testing.T) {
	session := &Session{}
	session.SetToken("my-token")

	if session.AuthToken() != "my-token" {
		t.Errorf("expected 'my-token', got %q", session.AuthToken())
	}
}

func TestUseBearerAuth_VersionThresholds(t *testing.T) {
	tests := []struct {
		major    int
		expected bool
	}{
		{4, false},
		{5, false},
		{6, false},
		{7, true},
		{8, true},
	}

	for _, tt := range tests {
		s := &Session{}
		s.ApiVersion.Major = tt.major
		if got := s.useBearerAuth(); got != tt.expected {
			t.Errorf("major=%d: expected %v, got %v", tt.major, tt.expected, got)
		}
	}
}

func TestNewSessionToken(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "7.0.0", nil
	})

	session, err := NewSessionToken(context.Background(), m.URL(), "my-api-token")
	if err != nil {
		t.Fatalf("NewSessionToken failed: %v", err)
	}
	if session.Token != "my-api-token" {
		t.Errorf("expected token 'my-api-token', got %q", session.Token)
	}
	if session.ApiVersion.Major != 7 {
		t.Errorf("expected Major=7, got %d", session.ApiVersion.Major)
	}
}

func TestDo_WithHTTPClient(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	customClient := &http.Client{Timeout: 5 * time.Second}
	session := &Session{
		URL:    m.URL(),
		Token:  "token",
		client: customClient,
	}
	session.ApiVersion.Major = 6

	m.handle("host.get", func(req *Request) (interface{}, *APIError) {
		return []interface{}{}, nil
	})

	_, err := session.GetHosts(context.Background(), HostGetParams{})
	if err != nil && !errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected error with custom http.Client: %v", err)
	}
}

func TestErrNotFound_IsSentinel(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should be comparable with errors.Is")
	}
}
