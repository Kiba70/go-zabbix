package zabbix

import (
	"context"
	"os"
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
