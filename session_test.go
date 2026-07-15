package zabbix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var session *Session

func GetTestCredentials() (username string, password string, url string) {
	url = os.Getenv("ZBX_URL")
	if url == "" {
		url = "http://localhost:8080/api_jsonrpc.php"
	}

	username = os.Getenv("ZBX_USERNAME")
	if username == "" {
		username = "Admin"
	}

	password = os.Getenv("ZBX_PASSWORD")
	if password == "" {
		password = "zabbix"
	}

	return username, password, url
}

func GetTestSession(t *testing.T) *Session {
	var err error
	if session == nil {
		username, password, url := GetTestCredentials()

		session, err = NewSession(context.Background(), url, username, password, "")
		if err != nil {
			t.Fatalf("Error creating a session: %v", err)
		}
	}

	return session
}

func TestSession(t *testing.T) {
	s := GetTestSession(t)

	v, err := s.GetVersion(context.Background())
	if err != nil || v == "" {
		t.Errorf("No API version found for session")
	}
}

func TestSessionDoRejectsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := &Session{URL: server.URL}
	_, err := s.Do(context.Background(), NewRequest("test", nil))
	if err == nil || !strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Fatalf("expected HTTP status error, got %v", err)
	}
}
