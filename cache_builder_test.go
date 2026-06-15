package zabbix

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------- SessionFileCache ----------

func newTempCacheFile(t *testing.T) (cache *SessionFileCache, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	cache = NewSessionFileCache().SetFilePath(path)
	cleanup = func() { os.RemoveAll(dir) }
	return cache, cleanup
}

func TestSessionFileCache_SaveLoad(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	original := &Session{
		URL:        "http://zabbix/api_jsonrpc.php",
		Token:      "abc123token",
		APIVersion: "6.0.0",
	}

	if err := cache.SaveSession(original); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if !cache.HasSession() {
		t.Fatal("HasSession should return true after save")
	}

	loaded, err := cache.GetSession()
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if loaded.URL != original.URL {
		t.Errorf("expected URL %q, got %q", original.URL, loaded.URL)
	}
	if loaded.Token != original.Token {
		t.Errorf("expected Token %q, got %q", original.Token, loaded.Token)
	}
	if loaded.APIVersion != original.APIVersion {
		t.Errorf("expected APIVersion %q, got %q", original.APIVersion, loaded.APIVersion)
	}
}

func TestSessionFileCache_HasSession_Empty(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	if cache.HasSession() {
		t.Fatal("HasSession should return false when no file exists")
	}
}

func TestSessionFileCache_GetSession_NoFile(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	_, err := cache.GetSession()
	if err == nil {
		t.Fatal("expected error when file doesn't exist, got nil")
	}
}

func TestSessionFileCache_Flush(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	original := &Session{URL: "http://test", Token: "tok", APIVersion: "1.0"}
	if err := cache.SaveSession(original); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if err := cache.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if cache.HasSession() {
		t.Fatal("HasSession should return false after Flush")
	}
}

func TestSessionFileCache_Flush_NoFile(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	// Flush on non-existent file should return an error.
	err := cache.Flush()
	if err == nil {
		t.Fatal("expected error when flushing non-existent file")
	}
}

func TestSessionFileCache_Expired(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	// Set lifetime to 1 second; wait > 1s for expiry.
	cache.SetSessionLifetime(1 * time.Second)

	original := &Session{URL: "http://test", Token: "tok"}
	if err := cache.SaveSession(original); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Wait for session to expire.
	time.Sleep(2 * time.Second)

	_, err := cache.GetSession()
	if err == nil {
		t.Fatal("expected error for expired session, got nil")
	}
}

func TestSessionFileCache_NotExpired(t *testing.T) {
	cache, cleanup := newTempCacheFile(t)
	defer cleanup()

	// Long lifetime.
	cache.SetSessionLifetime(1 * time.Hour)

	original := &Session{URL: "http://test", Token: "tok"}
	if err := cache.SaveSession(original); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := cache.GetSession()
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if loaded.Token != "tok" {
		t.Errorf("expected token 'tok', got %q", loaded.Token)
	}
}

func TestSessionFileCache_SetFilePath(t *testing.T) {
	cache := NewSessionFileCache()
	newPath := "/custom/path/session.json"
	returned := cache.SetFilePath(newPath)

	if cache.filePath != newPath {
		t.Errorf("expected filePath %q, got %q", newPath, cache.filePath)
	}
	// SetFilePath should return the cache for chaining.
	if returned != cache {
		t.Error("SetFilePath should return the cache instance for chaining")
	}
}

func TestSessionFileCache_SetFilePermissions(t *testing.T) {
	cache := NewSessionFileCache()
	returned := cache.SetFilePermissions(0644)

	if cache.filePermissions != 0644 {
		t.Errorf("expected 0644, got %v", cache.filePermissions)
	}
	if returned != cache {
		t.Error("SetFilePermissions should return the cache instance")
	}
}

func TestSessionFileCache_DefaultLifetime(t *testing.T) {
	cache := NewSessionFileCache()
	if cache.sessionLifeTime != 4*time.Hour {
		t.Errorf("expected default 4h, got %v", cache.sessionLifeTime)
	}
	if cache.filePermissions != 0600 {
		t.Errorf("expected default 0600, got %v", cache.filePermissions)
	}
}

// ---------- ClientBuilder ----------

func TestCreateClient_Defaults(t *testing.T) {
	builder := CreateClient("http://zabbix/api_jsonrpc.php")

	if builder.url != "http://zabbix/api_jsonrpc.php" {
		t.Errorf("unexpected url: %q", builder.url)
	}
	if builder.token != "" {
		t.Errorf("expected empty token, got %q", builder.token)
	}
	if builder.client == nil {
		t.Error("expected non-nil default http.Client")
	}
	if builder.credentials == nil {
		t.Error("expected non-nil credentials map")
	}
	if builder.hasCache {
		t.Error("expected hasCache=false by default")
	}
}

func TestClientBuilder_WithCredentials(t *testing.T) {
	builder := CreateClient("http://zabbix")
	builder.WithCredentials("admin", "pass")

	if builder.credentials["username"] != "admin" {
		t.Errorf("expected username 'admin', got %q", builder.credentials["username"])
	}
	if builder.credentials["password"] != "pass" {
		t.Errorf("expected password 'pass', got %q", builder.credentials["password"])
	}
}

func TestClientBuilder_WithToken(t *testing.T) {
	builder := CreateClient("http://zabbix")
	builder.WithToken("my-token")

	if builder.token != "my-token" {
		t.Errorf("expected token 'my-token', got %q", builder.token)
	}
}

func TestClientBuilder_WithHTTPClient(t *testing.T) {
	builder := CreateClient("http://zabbix")
	custom := &http.Client{}
	builder.WithHTTPClient(custom)

	if builder.client != custom {
		t.Error("expected custom http.Client to be set")
	}
}

func TestClientBuilder_WithCache(t *testing.T) {
	builder := CreateClient("http://zabbix")
	cache := NewSessionFileCache()
	builder.WithCache(cache)

	if !builder.hasCache {
		t.Error("expected hasCache=true")
	}
	if builder.cache == nil {
		t.Error("expected cache to be set")
	}
}

func TestClientBuilder_Connect_WithCache(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})

	// Pre-populate cache with a valid session.
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "cached.json")
	cache := NewSessionFileCache().SetFilePath(cachePath)

	cachedSession := &Session{
		URL:        m.URL(),
		Token:      "cached-token",
		APIVersion: "6.0.0",
	}
	cachedSession.ApiVersion.Major = 6

	if err := cache.SaveSession(cachedSession); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Connect should load from cache, not call login.
	session, err := CreateClient(m.URL()).
		WithCache(cache).
		WithCredentials("Admin", "zabbix").
		Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if session.Token != "cached-token" {
		t.Errorf("expected cached token 'cached-token', got %q", session.Token)
	}
}

func TestClientBuilder_Connect_WithToken(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "7.0.0", nil
	})

	session, err := CreateClient(m.URL()).
		WithToken("api-token-xyz").
		Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if session.Token != "api-token-xyz" {
		t.Errorf("expected token 'api-token-xyz', got %q", session.Token)
	}
	if session.ApiVersion.Major != 7 {
		t.Errorf("expected Major=7, got %d", session.ApiVersion.Major)
	}
}

func TestClientBuilder_Connect_WithCredentials(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return "login-token", nil
	})

	session, err := CreateClient(m.URL()).
		WithCredentials("Admin", "zabbix").
		Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if session.Token != "login-token" {
		t.Errorf("expected token 'login-token', got %q", session.Token)
	}
}

func TestClientBuilder_Connect_LoginError(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return nil, &APIError{Code: -1, Message: "Bad credentials"}
	})

	_, err := CreateClient(m.URL()).
		WithCredentials("wrong", "creds").
		Connect(context.Background())
	if err == nil {
		t.Fatal("expected error on login failure")
	}
}

func TestClientBuilder_Connect_SavesCache(t *testing.T) {
	m := newMockAPI(t)
	defer m.Close()

	m.handle("apiinfo.version", func(req *Request) (interface{}, *APIError) {
		return "6.0.0", nil
	})
	m.handle("user.login", func(req *Request) (interface{}, *APIError) {
		return "fresh-token", nil
	})

	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "new.json")
	cache := NewSessionFileCache().SetFilePath(cachePath)

	_, err := CreateClient(m.URL()).
		WithCache(cache).
		WithCredentials("Admin", "zabbix").
		Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Cache should now contain the session.
	if !cache.HasSession() {
		t.Fatal("expected session to be cached after Connect")
	}
}
